// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	libnet "github.com/shuLhan/share/lib/net"
)

// List of [ServerOptions] Debug mode.
// For example, to log DNS error and cache operations, set the Debug value
// to 3 or (DebugLevelDNS|DebugLevelCache).
const (
	// Log error on DNS level, in example EMPTY answer, ERR_NAME,
	// ERR_NOT_IMPLEMENTED, ERR_REFUSED.
	DebugLevelDNS = 1

	// Log cache operations, including new record, updating records,
	// and pruning record in caches.
	DebugLevelCache = 2

	// Log low level DNS connection and packet, including request and
	// response.
	DebugLevelConnPacket = 4
)

// ServerOptions describes options for running a DNS server.
type ServerOptions struct {
	primaryUDP []net.Addr // List of parent name server addresses using UDP.
	primaryTCP []net.Addr // List of parent name server addresses using TCP
	primaryDoh []string   // List of parent name server addresses using DoH.
	primaryDot []string   // List of parent name server addresses using DoT.

	ip net.IP

	// ListenAddress ip address and port number to serve query.
	// This field is optional, default to "0.0.0.0:53".
	ListenAddress string `ini:"dns:server:listen"`

	// TLSCertFile contains path to certificate for serving DNS over TLS
	// and HTTPS.
	// This field is optional, if its empty, server will listening on
	// unsecure HTTP connection only.
	TLSCertFile string `ini:"dns:server:tls.certificate"`

	// TLSPrivateKey contains path to certificate private key file.
	TLSPrivateKey string `ini:"dns:server:tls.private_key"`

	// NameServers contains list of parent name servers.
	//
	// Answer that does not exist on local will be forwarded to parent
	// name servers.  If this field is empty, any query that does not have
	// an answer in local caches, will be returned with response code
	// RCodeErrName (3).
	//
	// The name server use the URI format,
	//
	//	nameserver  = [ scheme "://" ] ( ip-address / domain-name ) [:port]
	//	scheme      = ( "udp" / "https" )
	//	ip-address  = ( ip4 / ip6 )
	//	domain-name = ; fully qualified domain name
	//
	// If no scheme is given, it will default to "udp".
	// The domain-name MUST only used if scheme is "https".
	//
	// Example,
	//
	//	udp://35.240.172.103
	//	https://35.240.172.103:853 (DNS over TLS)
	//	https://kilabit.info/dns-query (DNS over HTTPS)
	//
	NameServers []string `ini:"dns:server:parent"`

	// The root authority for all zones and records under this server.
	SOA RDataSOA

	// HTTPIdleTimeout number of seconds before considering the client of
	// HTTP connection to be closed.
	// This field is optional, default to 120 seconds.
	HTTPIdleTimeout time.Duration `ini:"dns:server:http.idle_timeout"`

	// PruneDelay define a delay where caches will be pruned.
	// This field is optional, minimum value is 1 minute, and default
	// value is 1 hour.
	// For example, if its set to 1 hour, every 1 hour the caches will be
	// inspected to remove answers that has not been accessed more than or
	// equal to PruneThreshold.
	PruneDelay time.Duration `ini:"dns:server:cache.prune_delay"`

	// PruneThreshold define negative duration where answers will be
	// pruned from caches.
	// This field is optional, minimum value is -1 minute, and default
	// value is -1 hour,
	// For example, if its set to -1 minute, any answers that has not been
	// accessed in the last 1 minute will be removed from cache.
	PruneThreshold time.Duration `ini:"dns:server:cache.prune_threshold"`

	// Debug level for server, accept value [DebugLevelDNS],
	// [DebugLevelCache], [DebugLevelConnPacket], or any combination of
	// it.
	Debug int `ini:"dns:server:debug"`

	// HTTPPort port for listening DNS over HTTP (DoH), default to 0.
	// If its zero, the server will not serve DNS over HTTP.
	HTTPPort uint16 `ini:"dns:server:http.port"`

	// TLSPort port for listening DNS over TLS, default to 0.
	// If its zero, the server will not serve DNS over TLS.
	TLSPort uint16 `ini:"dns:server:tls.port"`

	port uint16

	// TLSAllowInsecure option to allow to serve DoH with self-signed
	// certificate.
	// This field is optional.
	TLSAllowInsecure bool `ini:"dns:server:tls.allow_insecure"`

	// DoHBehindProxy allow serving DNS over insecure HTTP, even if
	// certificate file is defined.
	// This option allow serving DNS request forwarded by another proxy
	// server.
	DoHBehindProxy bool `ini:"dns:server:doh.behind_proxy"`
}

// init initialize the server options.
func (opts *ServerOptions) init() (err error) {
	opts.SOA.init()

	if len(opts.ListenAddress) == 0 {
		opts.ListenAddress = "0.0.0.0:53"
	}

	_, opts.ip, opts.port = libnet.ParseIPPort(opts.ListenAddress, DefaultPort)
	if opts.ip == nil {
		return fmt.Errorf("dns: invalid IP address '%s'", opts.ListenAddress)
	}

	if opts.HTTPIdleTimeout <= 0 {
		opts.HTTPIdleTimeout = defaultHTTPIdleTimeout
	}
	if opts.PruneDelay.Minutes() < 1 {
		opts.PruneDelay = time.Hour
	}
	if opts.PruneThreshold.Minutes() > -1 {
		opts.PruneThreshold = -1 * time.Hour
	}

	if len(opts.NameServers) == 0 {
		return nil
	}

	opts.initNameServers()

	if len(opts.primaryUDP) == 0 && len(opts.primaryDoh) == 0 && len(opts.primaryDot) == 0 {
		return fmt.Errorf("dns: no valid name servers")
	}

	return nil
}

func (opts *ServerOptions) getUDPAddress() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   opts.ip,
		Port: int(opts.port),
	}
}

func (opts *ServerOptions) getTCPAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   opts.ip,
		Port: int(opts.port),
	}
}

func (opts *ServerOptions) getHTTPAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   opts.ip,
		Port: int(opts.HTTPPort),
	}
}

func (opts *ServerOptions) getDoTAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   opts.ip,
		Port: int(opts.TLSPort),
	}
}

// parseNameServers parse each name server in NameServers list based on scheme
// and store the result either in udpAddrs, tcpAddrs, dohAddrs, or dotAddrs.
//
// If the name server format contains no scheme, it will be assumed to be
// "udp".
func (opts *ServerOptions) parseNameServers(nameServers []string) {
	var (
		dnsURL  *url.URL
		tcpAddr *net.TCPAddr
		udpAddr *net.UDPAddr
		ip      net.IP
		ns      string
		err     error
	)

	for _, ns = range nameServers {
		dnsURL, err = url.Parse(ns)
		if err != nil {
			log.Printf("dns: invalid name server URI %q", ns)
			continue
		}

		switch dnsURL.Scheme {
		case "tcp":
			tcpAddr, err = libnet.ParseTCPAddr(dnsURL.Host, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid IP address %q", dnsURL.Host)
				continue
			}
			opts.primaryTCP = append(opts.primaryTCP, tcpAddr)

		case "https":
			ip = net.ParseIP(dnsURL.Hostname())
			if ip == nil {
				opts.primaryDoh = append(opts.primaryDoh, ns)
			} else {
				opts.primaryDot = append(opts.primaryDot, dnsURL.Host)
			}

		default:
			if len(dnsURL.Host) > 0 {
				ns = dnsURL.Host
			}

			udpAddr, err = libnet.ParseUDPAddr(ns, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid IP address %q", ns)
				continue
			}
			opts.primaryUDP = append(opts.primaryUDP, udpAddr)

			tcpAddr, err = libnet.ParseTCPAddr(ns, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid IP address %q", ns)
				continue
			}
			opts.primaryTCP = append(opts.primaryTCP, tcpAddr)
		}
	}
}

func (opts *ServerOptions) initNameServers() {
	opts.primaryUDP = nil
	opts.primaryTCP = nil
	opts.primaryDoh = nil
	opts.primaryDot = nil
	opts.parseNameServers(opts.NameServers)
}
