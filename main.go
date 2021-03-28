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
		listenArgs      []string
		proxy           bool
		backendArg      string
		allowArgs       []string
		defaultHostname string
	)

	flag.Func("listen", "Socket to listen on (repeatable)", func(arg string) error {
		listenArgs = append(listenArgs, arg)
		return nil
	})
	flag.BoolVar(&proxy, "proxy", false, "Use PROXY protocol when talking to backend")
	flag.StringVar(&backendArg, "backend", "", ":PORT or /path/to/socket/dir for backends")
	flag.Func("allow", "CIDR of allowed backends (repeatable)", func(arg string) error {
		allowArgs = append(allowArgs, arg)
		return nil
	})
	flag.StringVar(&defaultHostname, "default-hostname", "", "Default hostname if client does not provide SNI")
	flag.Parse()

	server := &Server{
		ProxyProtocol:   proxy,
		DefaultHostname: defaultHostname,
	}

	if strings.HasPrefix(backendArg, "/") {
		server.Backend = &UnixDialer{Directory: backendArg}
	} else if strings.HasPrefix(backendArg, ":") {
		port := strings.TrimPrefix(backendArg, ":")
		if len(allowArgs) == 0 {
			log.Fatal("At least one -allow flag must be specified when you use TCP backends")
		}
		allowed, err := parseCIDRs(allowArgs)
		if err != nil {
			log.Fatal(err)
		}
		server.Backend = &TCPDialer{Port: port, Allowed: allowed}
	} else {
		log.Fatal("-backend must be a TCP port number (e.g. :443) or a path to a socket directory")
	}

	if len(listenArgs) == 0 {
		log.Fatal("At least one -listen flag must be specified")
	}

	listeners, err := listener.OpenAll(listenArgs)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.CloseAll(listeners)

	for _, l := range listeners {
		go serve(l, server)
	}

	select {}
}

func serve(listener net.Listener, server *Server) {
	log.Fatal(server.Serve(listener))
}
