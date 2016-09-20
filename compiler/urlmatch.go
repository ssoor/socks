package compiler

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

type JSONURLMatch struct {
	Host  string   `json:"host"`
	Url   string   `json:"url"`
	Match []string `json:"match"`
}

type matchData struct {
	matchs   []SMatch
	urlRegex *regexp.Regexp
}
type URLMatch struct {
	data map[string][]matchData
}

func NewURLMatch() *URLMatch {
	return &URLMatch{data: make(map[string][]matchData)}
}

func (sc *URLMatch) AddMatchs(jsonMatchs JSONURLMatch) (err error) {
	var urlmatch matchData

	if urlmatch.urlRegex, err = regexp.Compile(jsonMatchs.Url); err != nil {
		return err
	}

	for i := 0; i < len(jsonMatchs.Match); i++ {
		match, err := NewSMatch(jsonMatchs.Match[i])
		if err != nil {
			return err
		}

		urlmatch.matchs = append(urlmatch.matchs, match)
	}

	sc.data[jsonMatchs.Host] = append(sc.data[jsonMatchs.Host], urlmatch)
	return nil
}

func (sc *URLMatch) matchReplaces(md []matchData, url string, src string) (dst string, err error) {
	for _, urlmatch := range md {
		if false == urlmatch.urlRegex.MatchString(url) {
			continue
		}

		for _, match := range urlmatch.matchs {
			if dst, err := match.Replace(src); err == nil {
				return dst, nil
			}
		}
	}

	return src, errors.New("regular expression does not match")
}

func (sc *URLMatch) Replace(url *url.URL, src string) (dst string, err error) {
	host := strings.ToLower(url.Host)

	var exist bool
	var matchdatas []matchData

	matchdatas = sc.data[host] // 处理绝对匹配
	if dst, err = sc.matchReplaces(matchdatas, url.String(), src); nil == err {
		return
	}

	host = "." + host // 处理模糊匹配
	for i := 0; -1 != i; i = strings.IndexRune(host, '.') {
		host = host[i+1:]
		if matchdatas, exist = sc.data["."+host]; false == exist {
			continue
		}

		if dst, err = sc.matchReplaces(matchdatas, url.String(), src); nil == err {
			return
		}
	}

	matchdatas = sc.data["."] // 处理全局规则
	if dst, err = sc.matchReplaces(matchdatas, url.String(), src); nil == err {
		return
	}

	return src, errors.New("regular expression does not match")
}
