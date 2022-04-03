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
	"errors"
	"fmt"
	"io/fs"
	"net"
	"path/filepath"
)

type UnixDialer struct {
	Directory string
}

func (backend *UnixDialer) Dial(origHostname string, clientConn ClientConn) (BackendConn, error) {
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
