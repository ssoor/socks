package socks

import (
	"log"
	"net"
	"net/http"
	"os"
)

type HTTPTransport struct {
	http.Transport

	Rules *SRules
}

func init() {
	log.SetOutput(os.Stdout)
}

func NewHTTPTransport(forward Dialer, jsondata []byte) *HTTPTransport {

	transport := &HTTPTransport{
		Rules: NewSRules(forward),
		Transport: http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return forward.Dial(network, addr)
			},
		},
	}

	if err := transport.Rules.ResolveJson(jsondata); nil != err {
		log.Printf("Resolve json failed, err: %s\n", err)
	}

	return transport
}

func (this *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	tranpoort, resp := this.Rules.ResolveRequest(req)

	if nil != resp {
		return resp, nil
	}

	resp, err = tranpoort.RoundTrip(req)

	if err != nil {
		log.Printf("tranpoort round trip err: %s\n", err)
		return
	}

	resp = this.Rules.ResolveResponse(req, resp)

	return
}
