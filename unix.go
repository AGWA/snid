package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type UnixDialer struct {
	Directory string
}

func (backend *UnixDialer) Dial(origHostname string) (BackendConn, error) {
	hostname := backend.canonicalizeHostname(origHostname)
	if hostname == "" {
		return nil, fmt.Errorf("no backend found for %q", origHostname)
	}

	socketPath := filepath.Join(backend.hostnamePath(hostname), "socket")

	// TODO: consider setting a timeout on the dial
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Net: "unix", Name: socketPath})
	if err != nil {
		return nil, fmt.Errorf("dialing backend for host %q failed: %w", hostname, err)
	}

	return conn, nil
}

func (backend *UnixDialer) hostnamePath(hostname string) string {
	return filepath.Join(backend.Directory, hostname)
}

func (backend *UnixDialer) hostnameDirExists(hostname string) bool {
	hostnamePath := backend.hostnamePath(hostname)
	info, err := os.Stat(hostnamePath)
	if err == nil && info.IsDir() {
		return true
	}
	if err == nil {
		//log.Printf("Ignoring %s because it is not a directory", hostnamePath)
	} else if !os.IsNotExist(err) {
		//log.Printf("Ignoring %s due to stat error: %s", hostnamePath, err)
	}
	return false
}

func (backend *UnixDialer) canonicalizeHostname(hostname string) string {
	if len(hostname) == 0 || hostname[0] == '.' || strings.ContainsRune(hostname, '/') {
		return ""
	}

	hostname = strings.ToLower(hostname)
	hostname = strings.TrimRight(hostname, ".")

	if backend.hostnameDirExists(hostname) {
		return hostname
	}

	if wildcardHostname := replaceFirstLabel(hostname, "_"); backend.hostnameDirExists(wildcardHostname) {
		return wildcardHostname
	}

	return ""
}

func replaceFirstLabel(hostname string, replacement string) string {
	dot := strings.IndexByte(hostname, '.')
	if dot == -1 {
		return replacement
	} else {
		return replacement + hostname[dot:]
	}
}
