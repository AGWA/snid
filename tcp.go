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

	IPv6SourcePrefix net.IP
}

func (backend *TCPDialer) checkBackend(address string) error {
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

func (backend *TCPDialer) bindIPv6(sock syscall.RawConn, clientAddress net.Addr) error {
	clientTCPAddress, isTCP := clientAddress.(*net.TCPAddr)
	if !isTCP {
		return fmt.Errorf("client is not connected using TCP")
	}
	clientIPv4 := clientTCPAddress.IP.To4()
	if clientIPv4 == nil {
		return fmt.Errorf("client is not connected using IPv4")
	}
	sourceIPv6 := make(net.IP, 16)
	copy(sourceIPv6[:12], backend.IPv6SourcePrefix)
	copy(sourceIPv6[12:], clientIPv4)

	var controlErr error
	if err := sock.Control(func(fd uintptr) {
		controlErr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_FREEBIND, 1)
		if controlErr != nil {
			return
		}
		controlErr = syscall.Bind(int(fd), &syscall.SockaddrInet6{Addr: *(*[16]byte)(sourceIPv6)})
	}); err != nil {
		return err
	}
	return controlErr
}

func (backend *TCPDialer) network() string {
	if backend.IPv6SourcePrefix != nil {
		return "tcp6"
	} else {
		return "tcp"
	}
}

func (backend *TCPDialer) Dial(hostname string, clientAddress net.Addr) (BackendConn, error) {
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(network string, address string, c syscall.RawConn) error {
			if err := backend.checkBackend(address); err != nil {
				return err
			}
			if backend.IPv6SourcePrefix != nil {
				if err := backend.bindIPv6(c, clientAddress); err != nil {
					return err
				}
			}
			return nil
		},
	}

	conn, err := dialer.Dial(backend.network(), net.JoinHostPort(hostname, backend.Port))
	if err != nil {
		return nil, err
	}
	return conn.(*net.TCPConn), nil
}
