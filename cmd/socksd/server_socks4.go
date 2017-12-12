package main

import (
	"github.com/ssoor/socks/upstream"
	"net"

	"github.com/ssoor/socks"
)

func Socks4Serve(addr string, dialer socks.Dialer, decorators ...upstream.TransportDecorator) (error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	listener = upstream.NewDecorateListener(listener, decorators...)
	defer listener.Close()

	socks4Svr, err := socks.NewSocks4Server(dialer)
	
	if err != nil {
		return err
	}

	return socks4Svr.Serve(listener)
}