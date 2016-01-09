package socks

import (
	"errors"
	"net"
	"strconv"
)

// Socks5Client implements Socks5 Proxy Protocol(RFC 1928) Client Protocol.
// Just support CONNECT command, and support USERNAME/PASSWORD authentication methods(RFC 1929)
type Socks5Client struct {
	network  string
	address  string
	user     string
	password string
	forward  Dialer
}

// NewSocks5Client return a new Socks5Client that implements Dialer interface.
func NewSocks5Client(network, address, user, password string, forward Dialer) (*Socks5Client, error) {
	return &Socks5Client{
		network:  network,
		address:  address,
		user:     user,
		password: password,
		forward:  forward,
	}, nil
}

// Dial return a new net.Conn that through the CONNECT command to establish connections with proxy server.
// address as RFC's requirements that can be IPV4, IPV6 and domain host, such as 8.8.8.8:999 or google.com:80
func (s *Socks5Client) Dial(network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, errors.New("socks: no support for SOCKS5 proxy connections of type:" + network)
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

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	// check port
	port, err := strconv.Atoi(portStr)

	if err != nil {
		return nil, errors.New("socks: failed to parse port number: " + portStr)
	}

	if port < 1 || port > 0xffff {
		return nil, errors.New("socks: port number out of range: " + portStr)
	}

	if err := buildHandShakeRequest(conn, s.user, s.password); err != nil {
		return nil, err
	}

	if err := buildConnectionRequest(conn, host, port); err != nil {
		return nil, err
	}

	closeConn = nil
	return conn, nil
}
