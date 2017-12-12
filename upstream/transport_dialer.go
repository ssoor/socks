package upstream

import (
	"net"
	"time"

	"github.com/ssoor/socks"
)

type TransportDialer struct {
	dnsCache    *DNSCache
	timeoutDial time.Duration
}

func NewTransportDialer(timeoutDNSCache int, timeoutDial time.Duration) *TransportDialer {
	var dnsCache *DNSCache
	if timeoutDNSCache != 0 {
		dnsCache = NewDNSCache(timeoutDNSCache)
	}
	return &TransportDialer{
		dnsCache:    dnsCache,
		timeoutDial: time.Duration(timeoutDial) * time.Second,
	}
}

func parseAddress(address string) (interface{}, string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, "", err
	}
	ip := net.ParseIP(address)
	if ip != nil {
		return ip, port, nil
	} else {
		return host, port, nil
	}
}

func (d *TransportDialer) Dial(network, address string) (conn net.Conn, err error) {
	host, port, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	var dest string
	var ipCached bool
	switch h := host.(type) {
	case net.IP:
		{
			dest = h.String()
			ipCached = true
		}
	case string:
		{
			dest = h
			if d.dnsCache != nil {
				if p, ok := d.dnsCache.Get(h); ok {
					dest = p.String()
					ipCached = true
				}
			}
		}
	}
	address = net.JoinHostPort(dest, port)
	if 0 == d.timeoutDial {
		conn, err = socks.Direct.Dial(network, address)
	} else {
		conn, err = socks.Direct.DialTimeout(network, address, d.timeoutDial)
	}
	if err != nil {
		return nil, err
	}

	if d.dnsCache != nil && !ipCached {
		d.dnsCache.Set(host.(string), conn.RemoteAddr().(*net.TCPAddr).IP)
	}
	return conn, nil
}