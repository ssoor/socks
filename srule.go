package socks

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/ssoor/socks/compiler"
)

const (
	Rewrite = iota
	Redirect
)

type RuleTypeof int32

type jSONCompiler struct {
	Host  string   `json:"host"`
	Match []string `json:"match"`
}

type jSONSRule struct {
	Type RuleTypeof     `json:"type"`
	Host []jSONCompiler `json:"compilers"`
}

type JSONRules struct {
	Local  bool        `json:"local"`
	SRules []jSONSRule `json:"srules"`
}

type SRules struct {
	local    bool
	Rewrite  *compiler.SCompiler
	Redirect *compiler.SCompiler

	tranpoort_local  *http.Transport
	tranpoort_remote *http.Transport
}

func NewSRules(forward Dialer) *SRules {

	return &SRules{
		Rewrite:  compiler.NewSCompiler(),
		Redirect: compiler.NewSCompiler(),
		tranpoort_remote: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return forward.Dial(network, addr)
			},
		},
		tranpoort_local: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return Direct.Dial(network, addr)
			},
		},
	}
}

func (this *SRules) ResolveJson(data []byte) (err error) {

	jsonRules := JSONRules{}

	if err = json.Unmarshal(data, &jsonRules); err != nil {
		return err
	}

	for i := 0; i < len(jsonRules.SRules); i++ {
		for j := 0; j < len(jsonRules.SRules[i].Host); j++ {
			this.Add(jsonRules.SRules[i].Type, jsonRules.SRules[i].Host[j].Host, jsonRules.SRules[i].Host[j].Match)
		}
	}

	return nil
}

func (this *SRules) ResolveRequest(req *http.Request) (*http.Transport, *http.Response) {

	if dsturl, err := this.replaceURL(this.Redirect, req.Host, req.URL.String()); err == nil {
		if strings.EqualFold(req.URL.String(), dsturl.String()) {
			return this.tranpoort_remote, nil
		}

		log.Println("redirect: ", req.URL, " to ", dsturl)

		req.URL = dsturl

		return this.tranpoort_remote, this.createRedirectResponse(dsturl.String(), req)
	}

	if dsturl, err := this.replaceURL(this.Rewrite, req.Host, req.URL.String()); err == nil {

		if strings.EqualFold(req.URL.Host, dsturl.Host) {
			log.Println("rewrite: ", req.URL, " to ", dsturl)

			req.URL = dsturl

			return this.tranpoort_remote, nil
		} else {
			log.Println("rewrite err:", req.URL, " to ", dsturl)
		}
	}

	return this.tranpoort_local, nil
}

func (this *SRules) ResolveResponse(req *http.Request, resp *http.Response) *http.Response {

	if resp_type := resp.Header.Get("Content-Type"); strings.Contains(strings.ToLower(resp_type), "text/html") {
		log.Printf("text/html: %s\n", resp.Request.URL.String())
	}

	return resp
}

func (this *SRules) createRedirectResponse(url string, req *http.Request) (resp *http.Response) {

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

func (this *SRules) Add(typeof RuleTypeof, host string, rule []string) (err error) {

	switch typeof {
	case Rewrite:
		err = this.Rewrite.Add(host, rule)
	case Redirect:
		err = this.Redirect.Add(host, rule)
	default:
		return errors.New("无法是别的规则类型")
	}

	for i := 0; i < len(rule); i++ {
		log.Println(typeof, rule[i], err)
	}

	return err
}

func (this *SRules) replaceURL(scomp *compiler.SCompiler, host string, src string) (dsturl *url.URL, err error) {
	var dststr string

	if dststr, err = scomp.Replace(host, src); err != nil {
		return nil, err
	}

	if dsturl, err = url.Parse(dststr); err != nil {
		return nil, err
	}

	return dsturl, nil
}
