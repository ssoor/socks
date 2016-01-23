package main

import (
	"io"
	"net"
	"strings"
)

type CipherConn struct {
	net.Conn
	rwc io.ReadWriteCloser
}

func (this *CipherConn) Read(data []byte) (int, error) {
	return this.rwc.Read(data)
}

func (this *CipherConn) Write(data []byte) (int, error) {
	return this.rwc.Write(data)
}

func (this *CipherConn) Close() error {
	err := this.Conn.Close()
	this.rwc.Close()
	return err
}

func NewCipherConn(conn net.Conn, cryptMethod string, password string) (*CipherConn, error) {
	var rwc io.ReadWriteCloser
	var err error
	switch strings.ToLower(cryptMethod) {
	default:
		rwc = conn
	case "rc4":
		rwc, err = NewRC4Cipher(conn, []byte(password))
	case "des":
		rwc, err = NewDESCFBCipher(conn, []byte(password))
	case "aes-128-cfb":
		rwc, err = NewAESCFGCipher(conn, password, 16)
	case "aes-192-cfb":
		rwc, err = NewAESCFGCipher(conn, password, 24)
	case "aes-256-cfb":
		rwc, err = NewAESCFGCipher(conn, password, 32)
	}
	if err != nil {
		return nil, err
	}

	return &CipherConn{
		Conn: conn,
		rwc:  rwc,
	}, nil
}

func NewCipherConnDecorator(cryptoMethod, password string) ConnDecorator {
	return func(conn net.Conn) (net.Conn, error) {
		return NewCipherConn(conn, cryptoMethod, password)
	}
}
