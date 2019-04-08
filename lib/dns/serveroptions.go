// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
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
		return fmt.Errorf("dns: invalid address '%s'", opts.IPAddress)
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
