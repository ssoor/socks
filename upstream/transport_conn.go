package upstream

import (
	"net"

	"errors"

	"github.com/ssoor/socks"
)

var ErrorUpStaeamDail error = errors.New("forward connect service failed.")

type TransportDecorator func(net.Conn) (net.Conn, error)

type TransportConn struct {
	forward    socks.Dialer
	decorators []TransportDecorator
}

func decorateTransportConn(conn net.Conn, ds ...TransportDecorator) (net.Conn, error) {
	decorated := conn
	var err error
	for _, decorate := range ds {
		decorated, err = decorate(decorated)
		if err != nil {
			return nil, err
		}
	}
	return decorated, nil
}

func NewTransportConn(forward socks.Dialer, ds ...TransportDecorator) *TransportConn {
	d := &TransportConn{
		forward: forward,
	}
	d.decorators = append(d.decorators, ds...)
	return d
}

func (d *TransportConn) Dial(network, address string) (net.Conn, error) {
	conn, err := d.forward.Dial(network, address)
	if err != nil {
		return nil, err
	}
	dconn, err := decorateTransportConn(conn, d.decorators...)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return dconn, nil
}
