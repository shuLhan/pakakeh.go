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

//
// ServerOptions describes options for running a DNS server.
//
//nolint:maligned
type ServerOptions struct {
	// ListenAddress ip address and port number to serve query.
	// This field is optional, default to "0.0.0.0:53".
	ListenAddress string `ini:"dns:server:listen"`

	// HTTPIdleTimeout number of seconds before considering the client of
	// HTTP connection to be closed.
	// This field is optional, default to 120 seconds.
	HTTPIdleTimeout time.Duration `ini:"dns:server:http.idle_timeout"`

	// HTTPPort port for listening DNS over HTTP, default to 443.
	HTTPPort uint16 `ini:"dns:server:http.port"`

	// TLSPort port for listening DNS over TLS, default to 853.
	TLSPort uint16 `ini:"dns:server:tls.port"`

	//
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
	//	scheme      = ( "udp" / "tcp" / "https" )
	//	ip-address  = ( ip4 / ip6 )
	//	domain-name = ; fully qualified domain name
	//
	// If no scheme is given, it will default to "udp".
	// The domain-name MUST only used if scheme is "https".
	//
	// Example,
	//
	//	udp://35.240.172.103
	//	tcp://35.240.172.103:5353
	//	https://35.240.172.103:853 (DNS over TLS)
	//	https://kilabit.info/dns-query (DNS over HTTPS)
	//
	NameServers []string `ini:"dns:server:parent"`

	//
	// FallbackNS contains list of parent name servers that will be
	// queried if the primary NameServers return an error.
	//
	// This field use the same format as NameServers.
	//
	FallbackNS []string

	// TLSCertFile contains path to certificate for serving DNS over TLS
	// and HTTPS.
	// This field is optional, if its empty, server will listening on
	// unsecure HTTP connection only.
	TLSCertFile string `ini:"dns:server:tls.certificate"`

	// TLSPrivateKey contains path to certificate private key file.
	TLSPrivateKey string `ini:"dns:server:tls.private_key"`

	// TLSAllowInsecure option to allow to serve DoH with self-signed
	// certificate.
	// This field is optional.
	TLSAllowInsecure bool `ini:"dns:server:allow_insecure"`

	// DoHBehindProxy allow serving DNS over insecure HTTP, even if
	// certificate file is defined.
	// This option allow serving DNS request forwarded by another proxy
	// server.
	DoHBehindProxy bool `ini:"dns:server:doh.behind_proxy"`

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

	ip   net.IP
	port uint16

	// primaryUDP contains list of parent name server addresses using UDP
	// protocol.
	primaryUDP []*net.UDPAddr

	// primaryTCP contains list of parent name server addresses using TCP
	// protocol.
	primaryTCP []*net.TCPAddr

	// primaryDoh contains list of parent name server addresses using DoH
	// protocol.
	primaryDoh []string

	// primaryDoT contains list of parent name server addresses using DoT
	// protocol.
	primaryDoT []string

	fallbackUDP []*net.UDPAddr
	fallbackTCP []*net.TCPAddr
	fallbackDoh []string
}

//
// init initialize the server options.
//
func (opts *ServerOptions) init() (err error) {
	if len(opts.ListenAddress) == 0 {
		opts.ListenAddress = "0.0.0.0:53"
	}

	_, opts.ip, opts.port = libnet.ParseIPPort(opts.ListenAddress, DefaultPort)
	if opts.ip == nil {
		return fmt.Errorf("dns: invalid IP address '%s'", opts.ListenAddress)
	}

	if opts.HTTPPort == 0 {
		opts.HTTPPort = DefaultHTTPPort
	}
	if opts.TLSPort == 0 {
		opts.TLSPort = DefaultTLSPort
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

	opts.parseNameServers()

	if len(opts.primaryUDP) == 0 && len(opts.primaryTCP) == 0 && len(opts.primaryDoh) == 0 {
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

//
// parseNameServers parse each name server in NameServers list based on scheme
// and store the result either in udpAddrs, tcpAddrs, or dohAddrs.
//
// If the name server format contains no scheme, it will be assumed as "udp".
//
func parseNameServers(nameServers []string) (
	udpAddrs []*net.UDPAddr, tcpAddrs []*net.TCPAddr, dohAddrs []string,
) {
	for _, ns := range nameServers {
		dnsURL, err := url.Parse(ns)
		if err != nil {
			log.Printf("dns: invalid name server URI %q", ns)
			continue
		}

		switch dnsURL.Scheme {
		case "udp":
			udpAddr, err := libnet.ParseUDPAddr(dnsURL.Host, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid UDP IP address %q", dnsURL.Host)
				continue
			}

			udpAddrs = append(udpAddrs, udpAddr)

		case "tcp":
			tcpAddr, err := libnet.ParseTCPAddr(dnsURL.Host, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid TCP IP address %q", dnsURL.Host)
				continue
			}

			tcpAddrs = append(tcpAddrs, tcpAddr)

		case "https":
			dohAddrs = append(dohAddrs, ns)

		default:
			if len(dnsURL.Host) > 0 {
				ns = dnsURL.Host
			}

			udpAddr, err := libnet.ParseUDPAddr(ns, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid UDP IP address %q", ns)
				continue
			}

			udpAddrs = append(udpAddrs, udpAddr)
		}
	}

	return udpAddrs, tcpAddrs, dohAddrs
}

func (opts *ServerOptions) parseNameServers() {
	opts.primaryUDP, opts.primaryTCP, opts.primaryDoh = parseNameServers(opts.NameServers)
	opts.fallbackUDP, opts.fallbackTCP, opts.fallbackDoh = parseNameServers(opts.FallbackNS)
}
