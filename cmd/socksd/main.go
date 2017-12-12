package main

import (
	"time"
	"bytes"
	"flag"
	"net/http"

	"github.com/ssoor/socks"
	"github.com/ssoor/fundadore/log"
	"github.com/ssoor/socks/upstream"
)

func runHTTPProxy(conf Proxy, dialer socks.Dialer) {
	if conf.HTTP == "" {
		return
	}

	waitTime := float32(1)

	for {
		transport := NewHTTPTransport(dialer)

		if err := HTTPServe(conf.HTTP, dialer, transport); nil != err {
			ErrLog.Println("Start http proxy error: ", err)
		}

		waitTime += waitTime * 0.618
		log.Warning("http service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runSOCKS4Server(conf Proxy, dialer socks.Dialer) {
	if conf.SOCKS4 == "" {
		return
	}

	waitTime := float32(1)

	for {
		decorator := upstream.NewCipherDecorator(conf.Crypto, conf.Password)

		if err := Socks4Serve(conf.SOCKS4, dialer, decorator); nil != err {
			ErrLog.Println("Start socks4 proxy error: ", err)
		}

		waitTime += waitTime * 0.618
		log.Warning("http service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runSOCKS5Server(conf Proxy, dialer socks.Dialer) {
	if conf.SOCKS5 == "" {
		return
	}

	waitTime := float32(1)

	for {
		decorator := upstream.NewCipherDecorator(conf.Crypto, conf.Password)
		
		if err := Socks5Serve(conf.SOCKS5, dialer, decorator); nil != err {
			ErrLog.Println("Start socks5 proxy error: ", err)
		}

		waitTime += waitTime * 0.618
		log.Warning("http service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
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

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "socksd.json", "socksd start config info file path")

	flag.Parse()

	conf, err := LoadConfig(configFile)
	if err != nil {
		InfoLog.Printf("Load config: %s failed, err: %s\n", configFile, err)
		return
	}
	InfoLog.Printf("Load config: %s succeeded\n", configFile)

	for _, c := range conf.Proxies {
		upstreamSettings := upstream.Settings{
			DialTimeout: c.DNSCacheTimeout,
			DNSCacheTime: c.DNSCacheTimeout,
			Upstreams: c.Upstreams,
		}

		dialer := upstream.NewUpstreamDialer(upstreamSettings)

		go runHTTPProxy(c, dialer)
		go runSOCKS4Server(c, dialer)
		go runSOCKS5Server(c, dialer)
	}

	runPACServer(conf.PAC)
}
