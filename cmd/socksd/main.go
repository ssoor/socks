package main

import (
	"bytes"
	"errors"
	"flag"
	"net"
	"net/http"
	"strings"

	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"

	"github.com/ssoor/socks"
)

func Decrypt(base64Code []byte) (decode []byte, err error) {
	type encodeStruct struct {
		IV   string `json:"iv"`
		Code string `json:"code"`
	}

	key := []byte("890161F37139989CFA9433BAF32BDAFB")
	var jsonEninfo []byte

	for i := 0; i < len(base64Code); i++ {
		base64Code[i] = base64Code[i] - 0x90
	}

	if jsonEninfo, err = base64.StdEncoding.DecodeString(string(base64Code)); err != nil {
		return nil, err
	}

	eninfo := encodeStruct{}

	if err := json.Unmarshal(jsonEninfo, &eninfo); err != nil {
		return nil, err
	}

	var iv, code []byte

	iv, err = base64.StdEncoding.DecodeString(eninfo.IV)

	if err != nil {
		return nil, err
	}

	code, err = base64.StdEncoding.DecodeString(eninfo.Code)

	if err != nil {
		return nil, err
	}

	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(code, code)

	return code, nil
}

func GetValidByte(src []byte) []byte {
	var str_buf []byte
	for _, v := range src {
		if v != 0 {
			str_buf = append(str_buf, v)
		}
	}
	return str_buf
}

func getSRules(srcurl string) (jsonRules []byte, err error) {
	resp, err := http.Get(srcurl)

	if nil != err {
		return nil, err
	}

	defer resp.Body.Close()

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(resp.Body)

	jsonRules, err = Decrypt(bodyBuf.Bytes())

	if err != nil {
		return nil, err
	}

	return GetValidByte(jsonRules), nil
}

func main() {
	var isEncode bool
	var configFile string
	var userGUID string

	flag.BoolVar(&isEncode, "encode", true, "is start httpproxy encode to packets")

	flag.StringVar(&userGUID, "guid", "00000000_00000000", "is socksd start guid")
	flag.StringVar(&configFile, "config", "socksd.json", "socksd start config info file path")

	flag.Parse()

	conf, err := LoadConfig(configFile)
	if err != nil {
		InfoLog.Printf("Load config: %s failed, err: %s\n", configFile, err)
		return
	}
	InfoLog.Printf("Load config: %s succeeded\n", configFile)

	srules, err := getSRules("http://120.26.80.61/rule/issued/20160308/" + userGUID)
	if err != nil {
		InfoLog.Printf("Load srules: %s failed, err: %s\n", "http://angels.lingpao8.com/"+userGUID, err)
		return
	}
	InfoLog.Printf("Load srules: %s succeeded\n", "http://angels.lingpao8.com/"+userGUID)

	for _, c := range conf.Proxies {
		router := BuildUpstreamRouter(c)

		if isEncode {
			runHTTPEncodeProxyServer(c, router, srules)
		} else {
			runHTTPProxyServer(c, router, srules)
		}

		runSOCKS4Server(c, router)
		runSOCKS5Server(c, router)
	}

	/*
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Kill, os.Interrupt)

		<-sigChan
	*/

	runPACServer(conf.PAC)
}

func BuildUpstream(upstream Upstream, forward socks.Dialer) (socks.Dialer, error) {
	cipherDecorator := NewCipherConnDecorator(upstream.Crypto, upstream.Password)
	forward = NewDecorateClient(forward, cipherDecorator)

	switch strings.ToLower(upstream.Type) {
	case "socks5":
		{
			return socks.NewSocks5Client("tcp", upstream.Address, "", "", forward)
		}
	case "shadowsocks":
		{
			return socks.NewShadowSocksClient("tcp", upstream.Address, forward)
		}
	}
	return nil, errors.New("unknown upstream type" + upstream.Type)
}

func BuildUpstreamRouter(conf Proxy) socks.Dialer {
	var allForward []socks.Dialer
	for _, upstream := range conf.Upstreams {
		var forward socks.Dialer
		var err error
		forward = NewDecorateDirect(conf.DNSCacheTimeout)
		forward, err = BuildUpstream(upstream, forward)
		if err != nil {
			ErrLog.Println("failed to BuildUpstream, err:", err)
			continue
		}
		allForward = append(allForward, forward)
	}
	if len(allForward) == 0 {
		router := NewDecorateDirect(conf.DNSCacheTimeout)
		allForward = append(allForward, router)
	}
	return NewUpstreamDialer(allForward)
}

func runHTTPProxyServer(conf Proxy, router socks.Dialer, data []byte) {
	if conf.HTTP != "" {
		listener, err := net.Listen("tcp", conf.HTTP)
		if err != nil {
			ErrLog.Println("net.Listen at ", conf.HTTP, " failed, err:", err)
			return
		}
		go func() {
			defer listener.Close()
			httpProxy := socks.NewHTTPProxy(router, data)
			http.Serve(listener, httpProxy)
		}()
	}
}

func runHTTPEncodeProxyServer(conf Proxy, router socks.Dialer, data []byte) {
	if conf.HTTP != "" {
		listener, err := net.Listen("tcp", conf.HTTP)
		if err != nil {
			ErrLog.Println("net.Listen at ", conf.HTTP, " failed, err:", err)
			return
		}

		listener = socks.NewHTTPEncodeListener(listener)

		go func() {
			defer listener.Close()
			httpProxy := socks.NewHTTPProxy(router, data)
			http.Serve(listener, httpProxy)
		}()
	}
}

func runSOCKS4Server(conf Proxy, forward socks.Dialer) {
	if conf.SOCKS4 != "" {
		listener, err := net.Listen("tcp", conf.SOCKS4)
		if err != nil {
			ErrLog.Println("net.Listen failed, err:", err, conf.SOCKS4)
			return
		}
		cipherDecorator := NewCipherConnDecorator(conf.Crypto, conf.Password)
		listener = NewDecorateListener(listener, cipherDecorator)
		socks4Svr, err := socks.NewSocks4Server(forward)
		if err != nil {
			listener.Close()
			ErrLog.Println("socks.NewSocks4Server failed, err:", err)
		}
		go func() {
			defer listener.Close()
			socks4Svr.Serve(listener)
		}()
	}
}

func runSOCKS5Server(conf Proxy, forward socks.Dialer) {
	if conf.SOCKS5 != "" {
		listener, err := net.Listen("tcp", conf.SOCKS5)
		if err != nil {
			ErrLog.Println("net.Listen failed, err:", err, conf.SOCKS5)
			return
		}
		cipherDecorator := NewCipherConnDecorator(conf.Crypto, conf.Password)
		listener = NewDecorateListener(listener, cipherDecorator)
		socks5Svr, err := socks.NewSocks5Server(forward)
		if err != nil {
			listener.Close()
			ErrLog.Println("socks.NewSocks5Server failed, err:", err)
			return
		}
		go func() {
			defer listener.Close()
			socks5Svr.Serve(listener)
		}()
	}
}

func runPACServer(pac PAC) {
	pu, err := NewPACUpdater(pac)
	if err != nil {
		ErrLog.Println("failed to NewPACUpdater, err:", err)
		return
	}

	http.HandleFunc("/proxy.pac", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-ns-proxy-autoconfig")
		data, time := pu.get()
		reader := bytes.NewReader(data)
		http.ServeContent(w, r, "proxy.pac", time, reader)
	})

	err = http.ListenAndServe(pac.Address, nil)

	if err != nil {
		ErrLog.Println("listen failed, err:", err)
		return
	}
}
