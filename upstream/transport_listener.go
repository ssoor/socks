package upstream

import "net"

type TransportListener struct {
	listener   net.Listener
	decorators []TransportDecorator
}

func NewDecorateListener(listener net.Listener, ds ...TransportDecorator) *TransportListener {
	l := &TransportListener{
		listener: listener,
	}
	l.decorators = append(l.decorators, ds...)
	return l
}

func (s *TransportListener) Accept() (net.Conn, error) {
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}
	dconn, err := decorateTransportConn(conn, s.decorators...)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return dconn, nil
}

func (s *TransportListener) Close() error {
	return s.listener.Close()
}

func (s *TransportListener) Addr() net.Addr {
	return s.listener.Addr()
}
