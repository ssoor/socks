package socks

import (
	"errors"
	//ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"net"
	"strconv"
)

// ShadowSocksClient implements ShadowSocks Proxy Protocol
type ShadowSocksClient struct {
	network string
	address string
	forward Dialer
}

// NewShadowSocksClient return a new ShadowSocksClient that implements Dialer interface.
func NewShadowSocksClient(network, address string, forward Dialer) (*ShadowSocksClient, error) {
	return &ShadowSocksClient{
		network: network,
		address: address,
		forward: forward,
	}, nil
}

func RawAddr(host string, port int) (buf []byte, err error) {

	buff := make([]byte, 0, 1+(1+255)+2) // addrType + (lenByte + address) + port (host) | addrType + (address) + port (ip4 & ip6)

	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buff = append(buff, 1)
			ip = ip4
		} else {
			buff = append(buff, 4)
		}
		buff = append(buff, ip...)
	} else {
		if len(host) > 255 {
			return nil, errors.New("socks: destination hostname too long: " + host)
		}
		buff = append(buff, 3)
		buff = append(buff, uint8(len(host)))
		buff = append(buff, host...)
	}
	buff = append(buff, uint8(port>>8), uint8(port))

	return buff, nil
}

// Dial return a new net.Conn that through proxy server establish with address
func (s *ShadowSocksClient) Dial(network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, errors.New("socks: no support ShadowSocks proxy connections of type: " + network)
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("socks: failed to parse port number:" + portStr)
	}

	if port < 1 || port > 0xffff {
		return nil, errors.New("socks5: port number out of range:" + portStr)
	}

	conn, err := s.forward.Dial(s.network, s.address)
	if err != nil {
		return nil, err
	}

	closeConn := &conn

	defer func() {
		if closeConn != nil {
			(*closeConn).Close()
		}
	}()

	rawaddr, err := RawAddr(host, port)
	if err != nil {
		return nil, err
	}

	//remote, err = ss.DialWithRawAddr(rawaddr, address, se.cipher.Copy())

	_, err = conn.Write(rawaddr)
	if err != nil {
		return nil, err
	}

	closeConn = nil
	return conn, nil
}
