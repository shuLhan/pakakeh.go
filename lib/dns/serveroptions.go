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
// If certificate or key file is empty, server will not run with DNS over
// HTTPS (DoH).
//
type ServerOptions struct {
	// IPAddress ip address to serve query, without port number.
	// This field is optional, default to "0.0.0.0".
	IPAddress string

	// CertFile path to certificate file for serving DoH.
	// This field is optional.  If its defined, the DoHPrivateKeyFile must
	// also be defined.
	CertFile string

	// PrivateKeyFile path to certificate private key file for serving
	// DoH.
	// This field is optional.  If its defined, the CertFile must also
	// be defined.
	PrivateKeyFile string

	// DoHIdleTimeout number of seconds before considering the client of
	// DoH connection to be closed.
	// This field is optional, default to 120 seconds.
	DoHIdleTimeout time.Duration

	// UDPPort port for UDP server, default to 53.
	UDPPort uint16

	// TCPPort port for TCP server, default to 53.
	TCPPort uint16

	// DoHPort port for listening DNS over HTTP, default to 443.
	DoHPort uint16

	// NameServers contains list of parent name servers.
	//
	// Answer that does not exist on local will be forwarded to parent
	// name servers.  If this is empty, any query that does not have an
	// answer in local caches, will be returned with response code
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

	ip   net.IP
	cert *tls.Certificate

	// udpServers contains list of parent name server addresses using UDP
	// protocol.
	udpServers []*net.UDPAddr

	// tcpServers contains list of parent name server addresses using TCP
	// protocol.
	tcpServers []*net.TCPAddr

	// dohServers contains list of parent name server addresses using DoH
	// protocol.
	dohServers []string
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

	if len(opts.CertFile) > 0 && len(opts.PrivateKeyFile) > 0 {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.PrivateKeyFile)
		if err != nil {
			return fmt.Errorf("dns: error loading certificate: " + err.Error())
		}
		opts.cert = &cert
	}

	if opts.UDPPort == 0 {
		opts.UDPPort = DefaultPort
	}
	if opts.TCPPort == 0 {
		opts.TCPPort = DefaultPort
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

	if len(opts.udpServers) == 0 && len(opts.tcpServers) == 0 && len(opts.dohServers) == 0 {
		return fmt.Errorf("dns: no valid name servers")
	}

	return nil
}

func (opts *ServerOptions) getUDPAddress() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   opts.ip,
		Port: int(opts.UDPPort),
	}
}

func (opts *ServerOptions) getTCPAddress() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   opts.ip,
		Port: int(opts.TCPPort),
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
// and store the result either in udpServers, tcpServers, or dohServers.
// If the name server format is invalid, for example no scheme, it will be
// skipped.
//
func (opts *ServerOptions) parseNameServers() {
	for _, ns := range opts.NameServers {
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

			opts.udpServers = append(opts.udpServers, udpAddr)

		case "tcp":
			tcpAddr, err := libnet.ParseTCPAddr(dnsURL.Host, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid TCP IP address %q", dnsURL.Host)
				continue
			}

			opts.tcpServers = append(opts.tcpServers, tcpAddr)

		case "https":
			opts.dohServers = append(opts.dohServers, ns)

		default:
			udpAddr, err := libnet.ParseUDPAddr(dnsURL.Host, DefaultPort)
			if err != nil {
				log.Printf("dns: invalid UDP IP address %q", dnsURL.Host)
				continue
			}

			opts.udpServers = append(opts.udpServers, udpAddr)
		}
	}
}
