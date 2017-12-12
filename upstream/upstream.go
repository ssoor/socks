package upstream

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ssoor/socks"
	"github.com/ssoor/fundadore/log"
)

type UpstreamDialer struct {
	index   int
	lock    sync.Mutex
	dialers []socks.Dialer
}

func BuildUpstreamDialer(upstream Upstream, forward socks.Dialer) (socks.Dialer, error) {
	cipherDecorator := NewCipherDecorator(upstream.Crypto, upstream.Password)
	forward = NewTransportConn(forward, cipherDecorator)

	switch strings.ToLower(upstream.Type) {
	case "socks5":
		{
			return socks.NewSocks5Client("tcp", upstream.Address, "", "", forward)
		}
	case "shadowsocks":
		{
			return socks.NewShadowSocksClient("tcp", upstream.Address, forward)
		}
	}
	
	return nil, errors.New("unknown upstream type" + upstream.Type)
}

func buildSetting(setting Settings) []socks.Dialer {
	var err error
	var forward socks.Dialer
	var allForward []socks.Dialer

	for _, upstream := range setting.Upstreams {
		forward = NewTransportDialer(setting.DNSCacheTime, time.Duration(setting.DialTimeout))
		
		if forward, err = BuildUpstreamDialer(upstream, forward); nil != err {
			log.Warning("failed to BuildUpstream, err:", err)
		} else {
			allForward = append(allForward, forward)
		}
	}

	if len(allForward) == 0 {
		router := NewTransportDialer(setting.DNSCacheTime, time.Duration(setting.DialTimeout))
		allForward = append(allForward, router)
	}

	return allForward
}

func NewUpstreamDialer(setting Settings) *UpstreamDialer {
	dialer := &UpstreamDialer{
		index: 0,
		dialers:  buildSetting(setting), // 原始连接,不经过任何处理
	}

	return dialer
}

func (u *UpstreamDialer) CallNextDialer(network, address string) (conn net.Conn, err error) {
	u.lock.Lock()
	defer u.lock.Unlock()
	for {
		if 0 == len(u.dialers) {
			return socks.Direct.Dial(network, address)
		}

		if u.index++; u.index >= len(u.dialers) {
			u.index = 0
		}

		if conn, err = u.dialers[u.index].Dial(network, address); nil == err {
			break
		}

		switch err.(type) {
		case *net.OpError:
			if strings.EqualFold("dial", err.(*net.OpError).Op) {
				copy(u.dialers[u.index:], u.dialers[u.index+1:])
				u.dialers = u.dialers[:len(u.dialers)-1]

				log.Warning("Socks dial", network, address, "failed, delete current dialer:", err.(*net.OpError).Addr, ", err:", err)
				continue
			}
		}

		return nil, err
	}

	return conn, err
}

func (u *UpstreamDialer) Dial(network, address string) (conn net.Conn, err error) {
	if conn, err = u.CallNextDialer(network, address); nil != err {
		log.Warning("Upstream dial ", network, address, " failed, err:", err)
		return nil, err
	}

	return conn, nil
}
