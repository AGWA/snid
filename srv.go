package main

import (
	"fmt"
	"net"
	"strconv"
)

func getSRVService(protocols []string) string {
	if len(protocols) == 0 {
		return ""
	}
	switch protocols[0] {
	case "xmpp-client":
		return "xmpps-client"
	case "xmpp-server":
		return "xmpps-server"
	}
	return ""
}

func dialSRV(dialer net.Dialer, network string, hostname string, service string) (net.Conn, error) {
	_, addrs, err := net.LookupSRV(service, "tcp", hostname)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no SRV records exist for %s on %s", service, hostname)
	}

	var errs []error
	for _, addr := range addrs {
		conn, err := dialer.Dial(network, net.JoinHostPort(addr.Target, strconv.FormatUint(uint64(addr.Port), 10)))
		if err == nil {
			return conn, nil
		}
		errs = append(errs, err)
	}
	// TODO (Go 1.20): return nil, errors.Join(errs...)
	return nil, errs[0]
}
