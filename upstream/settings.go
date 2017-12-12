package upstream

type Upstream struct {
	Type     string `json:"type"`
	Crypto   string `json:"crypto"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

type Settings struct {
	DialTimeout  int        `json:"dial_timeout"`
	DNSCacheTime int        `json:"dnscache_time"`
	Upstreams    []Upstream `json:"services"`
}
