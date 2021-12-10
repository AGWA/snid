package main

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"src.agwa.name/go-listener/proxy"
	"src.agwa.name/go-listener/tlsutil"
)

type Server struct {
	Backend         BackendDialer
	ProxyProtocol   bool
	DefaultHostname string
}

func (server *Server) peekClientHello(clientConn net.Conn) (*tls.ClientHelloInfo, net.Conn, error) {
	if err := clientConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, nil, err
	}

	clientHello, peekedClientConn, err := tlsutil.PeekClientHelloFromConn(clientConn)
	if err != nil {
		return nil, nil, err
	}

	if err := clientConn.SetReadDeadline(time.Time{}); err != nil {
		return nil, nil, err
	}

	if clientHello.ServerName == "" {
		if server.DefaultHostname == "" {
			return nil, nil, errors.New("no SNI provided and DefaultHostname not set")
		}
		clientHello.ServerName = server.DefaultHostname
	}

	return clientHello, peekedClientConn, err
}

func (server *Server) handleConnection(clientConn net.Conn) {
	defer func() { clientConn.Close() }()

	var clientHello *tls.ClientHelloInfo

	if peekedClientHello, peekedClientConn, err := server.peekClientHello(clientConn); err == nil {
		clientHello = peekedClientHello
		clientConn = peekedClientConn
	} else {
		log.Printf("Peeking client hello from %s failed: %s", clientConn.RemoteAddr(), err)
		return
	}

	backendConn, err := server.Backend.Dial(clientHello.ServerName, clientConn.RemoteAddr())
	if err != nil {
		log.Printf("Ignoring connection from %s because dialing backend failed: %s", clientConn.RemoteAddr(), err)
		return
	}
	defer backendConn.Close()

	if server.ProxyProtocol {
		header := proxy.Header{RemoteAddr: clientConn.RemoteAddr(), LocalAddr: clientConn.LocalAddr()}
		if _, err := backendConn.Write(header.Format()); err != nil {
			log.Printf("Error writing PROXY header to backend: %s", err)
			return
		}
	}

	go func() {
		io.Copy(backendConn, clientConn)
		backendConn.CloseWrite()
	}()

	io.Copy(clientConn, backendConn)
}

func (server *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, isNetErr := err.(net.Error); isNetErr && netErr.Temporary() {
				log.Printf("Temporary network error accepting connection: %s", netErr)
				continue
			}
			return err
		}
		go server.handleConnection(conn)
	}
}
