package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"path/filepath"
)

type UnixDialer struct {
	Directory string
}

func (backend *UnixDialer) Dial(origHostname string, clientAddress net.Addr) (BackendConn, error) {
	hostname, err := canonicalizeHostname(origHostname)
	if err != nil {
		return nil, fmt.Errorf("invalid hostname %q", origHostname)
	}

	if conn, err := backend.dial(hostname); err == nil {
		return conn, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	if conn, err := backend.dial(wildcardHostname(hostname)); err == nil {
		return conn, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	return nil, fmt.Errorf("no backend socket found for %q", hostname)
}

func (backend *UnixDialer) dial(socketName string) (BackendConn, error) {
	socketPath := filepath.Join(backend.Directory, socketName)
	return net.DialUnix("unix", nil, &net.UnixAddr{Net: "unix", Name: socketPath})
}
