package main

import (
	"bytes"
	"encoding/json"
	"text/template"
)

var pacTemplate = `
{{range .Rules}}
var proxy_{{.Name}} = "{{if .Proxy}}PROXY {{.Proxy}};{{end}}{{if .SOCKS5}}SOCKS5 {{.SOCKS5}};{{end}}{{if .SOCKS4}}SOCKS4 {{.SOCKS4}};{{end}}DIRECT;";

var domains_{{.Name}} = {{.LocalRules}};
{{end}}

var regExpMatch = function(url, pattern) {
    try {
        return new RegExp(pattern).test(url); 
    } catch(ex) {
        return false; 
    }
};

var _internal_ScreeningURL4List_pos = 0;
var _internal_ScreeningURL4List_suffix = "";

var ScreeningURL4List = function(host, lists) {
    if (null == lists)
        return false;

    _internal_ScreeningURL4List_pos = host.lastIndexOf('.');

    while(1) {
        _internal_ScreeningURL4List_suffix = host.substring(_internal_ScreeningURL4List_pos + 1);

	    if (-1 != lists.indexOf(_internal_ScreeningURL4List_suffix)) {
	        return true;
	    }

        if (_internal_ScreeningURL4List_pos <= 0) {
            break;
        }

        _internal_ScreeningURL4List_pos = host.lastIndexOf('.', _internal_ScreeningURL4List_pos - 1);
    }

    return false;
}

var _internal_ScreeningURL4Shell_lastRule = '';

var ScreeningURL4Shell = function(url, shells) {
    if (null == shells)
        return false;
	
    for (var i = 0; i < shells.length; i++) {
    	_internal_ScreeningURL4Shell_lastRule = shells[i];

    	if(true == shExpMatch(url, _internal_ScreeningURL4Shell_lastRule)) {
    		return true;
    	}
    }

	return false;
}

var _internal_ScreeningURL4Regular_lastRule = '';

var ScreeningURL4Regular = function(url, regulars) {
    if (null == regulars)
        return false;
	
    for (var i = 0; i < regulars.length; i++) {
    	_internal_ScreeningURL4Regular_lastRule = regulars[i];
    	
    	if(true == regExpMatch(url, _internal_ScreeningURL4Regular_lastRule)) {
    		return true;
    	}
    }

	return false;
}

var ScreeningURL = function(url, host, filters) {
 
	if(true == ScreeningURL4List(host,filters["List"])){
		return true;
	}
 
	if(true == ScreeningURL4Shell(host,filters["Shell"])){
		return true;
	}
 
	if(true == ScreeningURL4Regular(host,filters["Regular"])){
		return true;
	}

	return false;
};

function FindProxyForURL(url, host) {
	if (isPlainHostName(host) || host === '127.0.0.1' || host === 'localhost') {
        return 'DIRECT';
    }

    {{range .Rules}}
	    if (ScreeningURL(url, host, domains_{{.Name}})) {
	        return proxy_{{.Name}};
	    }
	{{end}}

    return 'DIRECT';
}
`

type Expressions struct {
	List    []string
	Shell   []string
	Regular []string
}

type PACGenerator struct {
	Rules   []PACRule
	filters []Expressions
}

func NewPACGenerator(pacRules []PACRule) *PACGenerator {
	rules := make([]PACRule, len(pacRules))

	copy(rules, pacRules)

	return &PACGenerator{
		Rules:   rules,
		filters: make([]Expressions, len(rules)),
	}
}

func (this *PACGenerator) registrationList(index int, rules []string) error {
	this.filters[index].List = append(this.filters[index].List, rules...)

	return nil
}

func (this *PACGenerator) registrationShell(index int, rules []string) error {
	this.filters[index].Shell = append(this.filters[index].Shell, rules...)

	return nil
}

func (this *PACGenerator) registrationRegular(index int, rules []string) error {
	this.filters[index].Regular = append(this.filters[index].Shell, rules...)

	return nil
}

func (p *PACGenerator) Generate() ([]byte, error) {

	data := struct {
		Rules []PACRule
	}{
		Rules: p.Rules,
	}

	for i := 0; i < len(data.Rules); i++ {
		jsonstr, err := json.Marshal(p.filters[i])

		data.Rules[i].LocalRules = string(jsonstr)

		if err != nil {
			ErrLog.Println("failed to parse json, err:", err)
			return nil, err
		}
	}

	t, err := template.New("proxy.pac").Parse(pacTemplate)

	if err != nil {
		ErrLog.Println("failed to parse pacTempalte, err:", err)
		return nil, err
	}

	buff := bytes.NewBuffer(nil)
	err = t.Execute(buff, &data)
	if err != nil {
		InfoLog.Println(err)
		return nil, err
	}
	return buff.Bytes(), nil
}
