package compiler

import (
	"errors"
	"strings"
)

type SCompiler struct {
	matchs map[string][]SMatch
}

func NewSCompiler() *SCompiler {
	return &SCompiler{
		matchs: make(map[string][]SMatch),
	}
}

func (this *SCompiler) Add(host string, rule []string) error {
	host = strings.ToLower(host)

	for i := 0; i < len(rule); i++ {

		smatch, err := NewSMatch(rule[i])

		if err != nil {
			return err
		}

		this.matchs[host] = append(this.matchs[strings.ToLower(host)], smatch)
	}

	return nil
}

func (sc *SCompiler) matchReplaces(rules []SMatch, src string) (dst string, err error) {
	for _, match := range rules {
		if dst, err := match.Replace(src); err == nil {
			return dst, nil
		}
	}

	return src, errors.New("regular expression does not match")
}

func (sc *SCompiler) Replace(host string, src string) (dst string, err error) {
	host = strings.ToLower(host)

	var exist bool
	var rules []SMatch

	rules = sc.matchs[host] // 处理绝对匹配
	if dst, err = sc.matchReplaces(rules, src); nil == err {
		return
	}

	host = "." + host // 处理模糊匹配
	for i := 0; -1 != i; i = strings.IndexRune(host, '.') {
		host = host[i+1:]
		if rules, exist = sc.matchs["."+host]; false == exist {
			continue
		}

		if dst, err = sc.matchReplaces(rules, src); nil == err {
			return
		}
	}

	rules = sc.matchs["."] // 处理全局规则
	if dst, err = sc.matchReplaces(rules, src); nil == err {
		return
	}

	return src, errors.New("regular expression does not match")
}
