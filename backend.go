package main

import (
	"net"
)

type BackendConn interface {
	net.Conn
	CloseWrite() error
}

type ClientConn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type BackendDialer interface {
	Dial(string, ClientConn) (BackendConn, error)
}
