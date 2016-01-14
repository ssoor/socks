package compiler

import (
	"errors"
	"regexp"
	"strings"
)

type SMatch struct {
	template string
	regexp   *regexp.Regexp
}

func (this *SMatch) Init(rule string) error {

	if rule[0] != 's' && rule[0] != 'S' {
		return errors.New("invalid rule head: " + rule)
	}

	if rule[1] != '@' && rule[0] != '|' {
		return errors.New("invalid character segmentation: " + rule)
	}

	split := strings.Split(rule, rule[1:2])

	if len(split) != 4 {
		return errors.New("rule string incomplete or invalid: " + rule)
	}

	var err error
	this.regexp, err = regexp.Compile("(?" + split[3] + ")" + split[1])

	if err != nil {
		return err
	}

	this.template = split[2]

	return nil
}

func (this *SMatch) Replace(src string) (string, error) {
	var dst []byte

	submatch := this.regexp.FindStringSubmatchIndex(src)

	if len(submatch) == 0 {
		return src, errors.New("regular expression does not match")
	}

	return string(this.regexp.ExpandString(dst, this.template, src, submatch)), nil
}

type SCompiler struct {
	matchs map[string][]SMatch
}

func (this *SCompiler) Add(host string, rule string) error {

	smatch := SMatch{}

	err := smatch.Init(rule)

	if err != nil {
		return err
	}

	if this.matchs == nil {
		this.matchs = make(map[string][]SMatch)
	}

	this.matchs[host] = append(this.matchs[host], smatch)

	return nil
}

func (this *SCompiler) Replace(host string, src string) (dst string, err error) {
	rules := this.matchs[host]

	for _, match := range rules {
		if dst, err := match.Replace(src); err == nil {
			return dst, nil
		}
	}

	return src, errors.New("regular expression does not match")
}
