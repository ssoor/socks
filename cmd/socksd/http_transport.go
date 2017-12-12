package main

import (
	"github.com/ssoor/socks"
	"crypto/tls"
	"net"
	"io/ioutil"
	"net/http"
	"strings"

)

type HTTPTransport struct {
	tranpoort *http.Transport
}

func (this *HTTPTransport) create502Response(req *http.Request, err error) (resp *http.Response) {

	resp = &http.Response{
		StatusCode: http.StatusBadGateway,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header: http.Header{
			"X-Request-Error": []string{err.Error()},
		},
		ContentLength:    0,
		TransferEncoding: nil,
		Body:             ioutil.NopCloser(strings.NewReader("")),
		Close:            true,
	}

	return
}

func NewHTTPTransport(forward socks.Dialer) *HTTPTransport {
	transport := &HTTPTransport{
		tranpoort: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return forward.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return transport
}

func (h *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.Header.Del("X-Forwarded-For")

	if resp, err = h.tranpoort.RoundTrip(req); err != nil {
		if resp, err = h.tranpoort.RoundTrip(req); err != nil {
			return h.create502Response(req, err), nil
		}
	}

	return
}
