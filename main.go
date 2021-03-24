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
		backendPort     = flag.Int("backend-port", 443, "Port number of backend host")
		proxy           = flag.Bool("proxy", false, "Use PROXY protocol when talking to backend")
		defaultHostname = flag.String("default-hostname", "", "Default hostname if client does not provide SNI")
	)
	flag.Parse()

	allowCIDRs, err := parseCIDRs(strings.Split(*allow, ","))
	if err != nil {
		log.Fatal(err)
	}

	ourListeners, err := listener.OpenAll(flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	defer listener.CloseAll(ourListeners)

	server := &Server{
		AllowBackends:   allowCIDRs,
		BackendPort:     *backendPort,
		ProxyProtocol:   *proxy,
		DefaultHostname: *defaultHostname,
	}

	for _, listener := range ourListeners {
		go serve(listener, server)
	}

	select {}
}

func serve(listener net.Listener, server *Server) {
	log.Fatal(server.Serve(listener))
}
