package socks

import (
	"errors"
	"io"
	"net"
	"strconv"
)

const (
	socks5Version = 5

	socks5AuthNone     = 0
	socks5AuthPassword = 2
	socks5AuthNoAccept = 0xff

	socks5AuthPasswordVer = 1

	socks5Connect = 1

	socks5IP4    = 1
	socks5Domain = 3
	socks5IP6    = 4
)

const (
	socks5Success                 = 0
	socks5GeneralFailure          = 1
	socks5ConnectNotAllowed       = 2
	socks5NetworkUnreachable      = 3
	socks5HostUnreachable         = 4
	socks5ConnectionRefused       = 5
	socks5TTLExpired              = 6
	socks5CommandNotSupported     = 7
	socks5AddressTypeNotSupported = 8
)

var socks5Errors = []string{
	"",
	"general SOCKS server failure",
	"connection not allowed by ruleset",
	"network unreachable",
	"Host unreachable",
	"Connection refused",
	"TTL expired",
	"Command not supported",
	"Address type not supported",
}

func buildHandShakeRequest(conn net.Conn, user, password string) error {

	buff := make([]byte, 0, 3)

	buff = append(buff, socks5Version)

	// set authentication methods
	if len(user) > 0 && len(user) < 256 && len(password) < 256 {
		buff = append(buff, 2, socks5AuthNone, socks5AuthPassword)
	} else {
		buff = append(buff, 1, socks5AuthNone)
	}

	// send authentication methods
	if _, err := conn.Write(buff); err != nil {
		return errors.New("protocol: failed to write handshake request at: " + err.Error())
	}
	if _, err := io.ReadFull(conn, buff[:2]); err != nil {
		return errors.New("protocol: failed to read handshake reply at: " + err.Error())
	}

	// handle authentication methods reply
	if buff[0] != socks5Version {
		return errors.New("protocol: SOCKS5 server at: " + strconv.Itoa(int(buff[0])) + " invalid version")
	}
	if buff[1] == socks5AuthNoAccept {
		return errors.New("protocol: SOCKS5 server at: no acceptable methods")
	}

	if buff[1] == socks5AuthPassword {
		// build username/password authentication request
		buff = buff[:0]
		buff = append(buff, socks5AuthPasswordVer)
		buff = append(buff, uint8(len(user)))
		buff = append(buff, []byte(user)...)
		buff = append(buff, uint8(len(password)))
		buff = append(buff, []byte(password)...)

		if _, err := conn.Write(buff); err != nil {
			return errors.New("protocol: failed to write password authentication request to SOCKS5 server at: " + err.Error())
		}
		if _, err := io.ReadFull(conn, buff[:2]); err != nil {
			return errors.New("protocol: failed to read password authentication reply from SOCKS5 server at: " + err.Error())
		}
		// 0 indicates success
		if buff[1] != 0 {
			return errors.New("protocol: SOCKS5 server at: reject username/password")
		}
	}

	return nil
}

func buildConnectionRequest(conn net.Conn, host string, port int) error {

	buff := make([]byte, 0, 6+len(host))

	// build connect request
	buff = buff[:0]
	buff = append(buff, socks5Version, socks5Connect, 0)

	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buff = append(buff, socks5IP4)
			ip = ip4
		} else {
			buff = append(buff, socks5IP6)
		}
		buff = append(buff, ip...)
	} else {
		if len(host) > 255 {
			return errors.New("protocol: destination hostname too long: " + host)
		}
		buff = append(buff, socks5Domain)
		buff = append(buff, uint8(len(host)))
		buff = append(buff, host...)
	}
	buff = append(buff, byte(port>>8), byte(port))

	if _, err := conn.Write(buff); err != nil {
		return errors.New("protocol: failed to write connect request to SOCKS5 server at: " + err.Error())
	}
	if _, err := io.ReadFull(conn, buff[:4]); err != nil {
		return errors.New("protocol: failed to read connect reply from SOCKS5 server at: " + err.Error())
	}

	failure := "Undefined REP field"
	if int(buff[1]) < len(socks5Errors) {
		failure = socks5Errors[buff[1]]
	}
	if len(failure) > 0 {
		return errors.New("protocol: SOCKS5 server failed to connect: " + failure)
	}

	// read remain data include BIND.ADDRESS and BIND.PORT
	discardBytes := 0
	switch buff[3] {
	case socks5IP4:
		discardBytes = net.IPv4len
	case socks5IP6:
		discardBytes = net.IPv6len
	case socks5Domain:
		if _, err := io.ReadFull(conn, buff[:1]); err != nil {
			return errors.New("protocol: failed to read domain length from SOCKS5 server at: " + err.Error())
		}
		discardBytes = int(buff[0])
	default:
		return errors.New("protocol: got unknown address type " + strconv.Itoa(int(buff[3])) + " from SOCKS5 server")
	}
	discardBytes += 2
	if cap(buff) < discardBytes {
		buff = make([]byte, discardBytes)
	} else {
		buff = buff[:discardBytes]
	}
	if _, err := io.ReadFull(conn, buff); err != nil {
		return errors.New("protocol: failed to read address and port from SOCKS5 server at: " + err.Error())
	}

	return nil
}
