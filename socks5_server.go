package socks

import (
	"io"
	"net"
	"strconv"
)

// Socks5Server implements Socks5 Proxy Protocol(RFC 1928), just support CONNECT command.
type Socks5Server struct {
	forward Dialer
}

// NewSocks5Server return a new Socks5Server
func NewSocks5Server(forward Dialer) (*Socks5Server, error) {
	return &Socks5Server{
		forward: forward,
	}, nil
}

func serveSocks5Client(conn net.Conn, forward Dialer) {
	defer conn.Close()

	buff := make([]byte, 262)
	reply := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x22, 0x22}

	if _, err := io.ReadFull(conn, buff[:2]); err != nil {
		return
	}
	if buff[0] != socks5Version {
		reply[1] = socks5AuthNoAccept
		conn.Write(reply[:2])
		return
	}
	numMethod := buff[1]
	if _, err := io.ReadFull(conn, buff[:numMethod]); err != nil {
		return
	}
	reply[1] = socks5AuthNone
	if _, err := conn.Write(reply[:2]); err != nil {
		return
	}

	if _, err := io.ReadFull(conn, buff[:4]); err != nil {
		return
	}
	if buff[1] != socks5Connect {
		reply[1] = socks5CommandNotSupported
		conn.Write(reply)
		return
	}

	addressType := buff[3]
	addressLen := 0
	switch addressType {
	case socks5IP4:
		addressLen = net.IPv4len
	case socks5IP6:
		addressLen = net.IPv6len
	case socks5Domain:
		if _, err := io.ReadFull(conn, buff[:1]); err != nil {
			return
		}
		addressLen = int(buff[0])
	default:
		reply[1] = socks5AddressTypeNotSupported
		conn.Write(reply)
		return
	}
	host := make([]byte, addressLen)
	if _, err := io.ReadFull(conn, host); err != nil {
		return
	}
	if _, err := io.ReadFull(conn, buff[:2]); err != nil {
		return
	}
	hostStr := ""
	switch addressType {
	case socks5IP4, socks5IP6:
		ip := net.IP(host)
		hostStr = ip.String()
	case socks5Domain:
		hostStr = string(host)
	}
	port := uint16(buff[0])<<8 | uint16(buff[1])
	if port < 1 || port > 0xffff {
		reply[1] = socks5HostUnreachable
		conn.Write(reply)
		return
	}
	portStr := strconv.Itoa(int(port))

	hostStr = net.JoinHostPort(hostStr, portStr)
	dest, err := forward.Dial("tcp", hostStr)
	if err != nil {
		reply[1] = socks5ConnectionRefused
		conn.Write(reply)
		return
	}
	defer dest.Close()
	reply[1] = socks5Success
	if _, err := conn.Write(reply); err != nil {
		return
	}

	go func() {
		defer conn.Close()
		defer dest.Close()
		io.Copy(conn, dest)
	}()

	io.Copy(dest, conn)
}

// Serve with net.Listener for new incoming clients.
func (s *Socks5Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return err
			}
		}

		go serveSocks5Client(conn, s.forward)
	}
}
