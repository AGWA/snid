package main

import (
	"flag"
	"log"
	"net"
	"strings"

	"src.agwa.name/go-listener"
)

func parseCIDRs(strings []string) ([]*net.IPNet, error) {
	cidrs := make([]*net.IPNet, len(strings))
	for i := range strings {
		var err error
		_, cidrs[i], err = net.ParseCIDR(strings[i])
		if err != nil {
			return nil, err
		}
	}
	return cidrs, nil
}

func main() {
	var (
		allow           = flag.String("allow", "", "Comma-separated list of CIDRs to allow")
		backend         = flag.String("backend", "", ":PORT or /path/to/socket/dir for backends")
		proxy           = flag.Bool("proxy", false, "Use PROXY protocol when talking to backend")
		defaultHostname = flag.String("default-hostname", "", "Default hostname if client does not provide SNI")
	)
	flag.Parse()

	server := &Server{
		ProxyProtocol:   *proxy,
		DefaultHostname: *defaultHostname,
	}

	if strings.HasPrefix(*backend, "/") {
		server.Backend = &UnixDialer{Directory: *backend}
	} else if strings.HasPrefix(*backend, ":") {
		port := strings.TrimPrefix(*backend, ":")
		if *allow == "" {
			log.Fatal("-allow must be specified when you use TCP backends")
		}
		allowedCIDRs, err := parseCIDRs(strings.Split(*allow, ","))
		if err != nil {
			log.Fatal(err)
		}
		server.Backend = &TCPDialer{Port: port, Allowed: allowedCIDRs}
	} else {
		log.Fatal("-backend must be a TCP port number (e.g. :443) or a path to a socket directory")
	}

	if flag.NArg() == 0 {
		log.Fatal("At least one listener must be specified on the command line")
	}

	ourListeners, err := listener.OpenAll(flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	defer listener.CloseAll(ourListeners)

	for _, listener := range ourListeners {
		go serve(listener, server)
	}

	select {}
}

func serve(listener net.Listener, server *Server) {
	log.Fatal(server.Serve(listener))
}
