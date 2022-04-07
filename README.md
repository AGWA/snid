# snid - SNI-based Proxy Server

snid is a lightweight proxy server that forwards TLS connections based on the server name indication (SNI) hostname.  snid favors convention over configuration - backend addresses are not configured individually, but rather constructed based on DNS record lookups or filesystem locations.  This makes snid deployments easy to manage.

## Command Line Arguments

### `-listen LISTENER` (Mandatory)

Listen on the given address, provided in [go-listener syntax](https://github.com/AGWA/go-listener#listener-syntax).  You can specify the `-listen` flag multiple times to listen on multiple addresses.

Examples:
* `-listen tcp:443` to listen on TCP port 443, all interfaces.
* `-listen tcp:192.0.2.4:443` to listen on TCP port 443 on 192.0.2.4.

### `-mode nat46`, `-mode tcp`, or `-mode unix` (Mandatory)

Use the given mode, described below.

### `-default-hostname HOSTNAME` (Optional)

Use the given hostname if a client does not include the SNI extension.  If this flag is not specified, then SNI-less connections will be terminated with a TLS alert.


## NAT46 mode

In NAT46 mode, snid does an AAAA record lookup on the SNI hostname and forwards the connection there, as long as the IPv6 address is within one of the networks specified by `-backend-cidr`.  The client's IPv4 address is embedded in the lower 4 bytes of the source address used for connecting to the backend, with the prefix specified by `-nat46-prefix`.

Note: in NAT46 mode, clients which connect to snid over IPv6 will be disconnected. Instead, IPv6 clients should connect directly to the backend.

The following flags can be specified in NAT46 mode:

### `-nat46-prefix IPV6ADDRESS` (Mandatory)

Use the given prefix for the source address when connecting to the backend.  Specifically, the source address is constructed by taking the IPv6 address specified by `-nat46-prefix` and placing the client's IPv4 address in the lower 4 bytes.

It is recommended that you use one of the prefixes reserved by [RFC 8215](https://datatracker.ietf.org/doc/html/rfc8215) for IPv4/IPv6 translation mechanisms, such as `64:ff9b:1::`.

Example: `-nat46-prefix 64:ff9b:1::`

Important: the prefix which you use for `-nat46-prefix` MUST be routed to the local host so that return packets can reach snid.  On Linux, the necessary route entry can be added by running:

```
ip route add local 64:ff9b:1::/96 dev lo
```

### `-backend-cidr CIDR` (Mandatory)

Only forward connections to addresses within the given subnet.  This option can be specified multiple times to allow multiple subnets.

Example: `-backend-cidr 2001:db8::/64`


## TCP mode

In TCP mode, snid does an A/AAAA record lookup on the SNI hostname and forwards the connection there, as long as the IP address is within one of the networks specified by `-backend-cidr`.

The following flags can be specified in TCP mode:

### `-backend-cidr CIDR` (Mandatory)

Only forward connections to addresses within the given subnet.  This option can be specified multiple times to allow multiple subnets.

Examples:
* `-backend-cidr 192.0.2.0/24`
* `-backend-cidr 2001:db8::/64`

### `-backend-port PORTNO` (Optional)

Connect to the given port number on the backend.

If this option is omitted, then snid will use the same port number that the inbound connection arrived on.

### `-proxy-proto` (Optional)

Use [PROXY protocol v2](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt) to convey the client IP address to the backend.


## UNIX mode

In UNIX mode, snid forwards connections to a UNIX domain socket whose filename is the SNI hostname, in the directory specified by `-unix-directory`.

The following flags can be specified with UNIX mode:

### `-unix-directory PATH` (Mandatory)

The path to the directory containing UNIX domain sockets.

### `-proxy-proto` (Optional)

Use [PROXY protocol v2](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt) to convey the client IP address to the backend.
