package socks

import (
	"fmt"
	"github.com/ssoor/socks/compiler"
	"io/ioutil"
	"net/http"
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

var rules compiler.SCompiler

func init() {

	rules.Add("mail.yeah.net", "s@^(http[s]?)://[^/]*[/]*/\\?(?:kb02464p|751)@$1://1.sogoulp.com/index5883_1.html@i")

}

func (t *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	srcUrl := req.URL.String()

	if dstUrl, err := rules.Replace(req.Host, srcUrl); err == nil {

		fmt.Printf("%s(%s) == %s\n", req.Host, srcUrl, dstUrl)

		resp = createRedirectResponse(dstUrl, req)

		return resp, nil
	}

	resp, err = t.Transport.RoundTrip(req)

	if err != nil {
		return
	}

	return
}
