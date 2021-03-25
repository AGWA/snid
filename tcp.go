package main

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

type TCPDialer struct {
	Port    string
	Allowed []*net.IPNet
}

func (backend *TCPDialer) dialControl(network string, address string, c syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ipaddress := net.ParseIP(host)
	if ipaddress == nil {
		return fmt.Errorf("%s is not a valid IP address", host)
	}
	for _, cidr := range backend.Allowed {
		if cidr.Contains(ipaddress) {
			return nil
		}
	}
	return fmt.Errorf("%s is not an allowed backend", ipaddress)
}

func (backend *TCPDialer) Dial(hostname string) (BackendConn, error) {
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
		Control: backend.dialControl,
	}

	conn, err := dialer.Dial("tcp", net.JoinHostPort(hostname, backend.Port))
	if err != nil {
		return nil, err
	}
	return conn.(*net.TCPConn), nil
}
