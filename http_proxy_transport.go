package socks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
)

type HTTPTransport struct {
	http.Transport
}

func createRedirectResponse(url string, req *http.Request) (resp *http.Response) {

	resp = &http.Response{
		StatusCode: http.StatusFound,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header: http.Header{
			"Location": []string{url},
		},
		ContentLength:    0,
		TransferEncoding: nil,
		Body:             ioutil.NopCloser(strings.NewReader("")),
		Close:            true,
	}

	return
}

func (t *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	if strings.EqualFold(strings.ToLower(req.URL.String()), "http://0.baidu.com/") {

		resp = createRedirectResponse("http://www.baidu.com", req)

		response, _ := httputil.DumpResponse(resp, true)

		fmt.Println(string(response))
	} else {
		resp, err = t.Transport.RoundTrip(req)
		if err != nil {
			return
		}
	}
	return
}
