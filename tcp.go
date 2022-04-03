// Copyright (C) 2022 Andrew Ayer
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// Except as contained in this notice, the name(s) of the above copyright
// holders shall not be used in advertising or otherwise to promote the
// sale, use or other dealings in this Software without prior written
// authorization.

package main

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"
)

type TCPDialer struct {
	Port    int
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

func (backend *TCPDialer) bindIPv6(sock syscall.RawConn, clientConn ClientConn) error {
	clientTCPAddress, isTCP := clientConn.RemoteAddr().(*net.TCPAddr)
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

func (backend *TCPDialer) port(clientConn ClientConn) (int, error) {
	if backend.Port != 0 {
		return backend.Port, nil
	}

	localTCPAddress, isTCP := clientConn.LocalAddr().(*net.TCPAddr)
	if !isTCP {
		return 0, fmt.Errorf("cannot determine backend port number because client is not connected using TCP")
	}
	return localTCPAddress.Port, nil
}

func (backend *TCPDialer) network() string {
	if backend.IPv6SourcePrefix != nil {
		return "tcp6"
	} else {
		return "tcp"
	}
}

func (backend *TCPDialer) Dial(hostname string, clientConn ClientConn) (BackendConn, error) {
	port, err := backend.port(clientConn)
	if err != nil {
		return nil, err
	}

	dialer := net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(network string, address string, c syscall.RawConn) error {
			if err := backend.checkBackend(address); err != nil {
				return err
			}
			if backend.IPv6SourcePrefix != nil {
				if err := backend.bindIPv6(c, clientConn); err != nil {
					return err
				}
			}
			return nil
		},
	}

	conn, err := dialer.Dial(backend.network(), net.JoinHostPort(hostname, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return conn.(*net.TCPConn), nil
}
