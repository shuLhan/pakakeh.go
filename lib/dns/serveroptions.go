// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
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
	// IPAddress ip address to serve query, without port number.
	// This field is optional, default to "0.0.0.0".
	IPAddress string

	// DoHIdleTimeout number of seconds before considering the client of
	// DoH connection to be closed.
	// This field is optional, default to 120 seconds.
	DoHIdleTimeout time.Duration

	// Port port for UDP and TCP server, default to 53.
	Port uint16

	// DoHPort port for listening DNS over HTTP, default to 443.
	DoHPort uint16

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
	//	udp://1.1.1.1
	//	tcp://192.168.1.1:5353
	//	https://cloudflare-dns.com/dns-query
	//
	NameServers []string

	//
	// FallbackNS contains list of parent name servers that will be
	// queried if the primary NameServers return an error.
	//
	// This field use the same format as NameServers.
	//
	FallbackNS []string

	// DoHCertificate contains certificate for serving DNS over HTTPS.
	// This field is optional, if its empty, server will not listening on
	// HTTPS port.
	DoHCertificate *tls.Certificate

	// DoHAllowInsecure option to allow to serve DoH with self-signed
	// certificate.
	// This field is optional.
	DoHAllowInsecure bool

	// PruneDelay define a delay where caches will be pruned.
	// This field is optional, minimum value is 1 minute, and default
	// value is 1 hour.
	// For example, if its set to 1 hour, every 1 hour the caches will be
	// inspected to remove answers that has not been accessed more than or
	// equal to PruneThreshold.
	PruneDelay time.Duration

	// PruneThreshold define negative duration where answers will be
	// pruned from caches.
	// This field is optional, minimum value is -1 minute, and default
	// value is -1 hour,
	// For example, if its set to -1 minute, any answers that has not been
	// accessed in the last 1 minute will be removed from cache.
	PruneThreshold time.Duration

	ip net.IP

	// primaryUDP contains list of parent name server addresses using UDP
	// protocol.
	primaryUDP []*net.UDPAddr

	// primaryTCP contains list of parent name server addresses using TCP
	// protocol.
	primaryTCP []*net.TCPAddr

	// primaryDoh contains list of parent name server addresses using DoH
	// protocol.
	primaryDoh []string

	fallbackUDP []*net.UDPAddr
	fallbackTCP []*net.TCPAddr
	fallbackDoh []string
}

//
// init initialize the server options.
//
func (opts *ServerOptions) init() (err error) {
	if len(opts.IPAddress) == 0 {
		opts.IPAddress = "0.0.0.0"
	}

	opts.ip = net.ParseIP(opts.IPAddress)
	if opts.ip == nil {
		return fmt.Errorf("dns: invalid IP address '%s'", opts.IPAddress)
	}

	if opts.Port == 0 {
		opts.Port = DefaultPort
	}
	if opts.DoHPort == 0 {
		opts.DoHPort = DefaultDoHPort
	}
	if opts.DoHIdleTimeout <= 0 {
		opts.DoHIdleTimeout = defaultDoHIdleTimeout
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
		Port: int(opts.Port),
	}
}

func (opts *ServerOptions) getTCPAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   opts.ip,
		Port: int(opts.Port),
	}
}

func (opts *ServerOptions) getDoHAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   opts.ip,
		Port: int(opts.DoHPort),
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
