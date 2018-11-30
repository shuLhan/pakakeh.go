// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
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
	// IPAddress of server to listen to, without port number.
	IPAddress string

	// DoHCert path to certificate file for serving DoH.
	DoHCert string

	// DoHCertKey path to certificate key file for serving DoH.
	DoHCertKey string

	// DoHIdleTimeout number of seconds before considering the client of
	// DoH connection to be closed.
	DoHIdleTimeout time.Duration

	ip net.IP

	// UDPPort port for UDP server, default to 53.
	UDPPort uint16

	// TCPPort port for TCP server, default to 53.
	TCPPort uint16

	// DoHPort port for listening DNS over HTTP, default to 443.
	DoHPort uint16

	// DoHAllowInsecure options to allow to serve DoH with self-signed
	// certificate.
	DoHAllowInsecure bool
}

func (opts *ServerOptions) parse() error {
	ip := net.ParseIP(opts.IPAddress)
	if ip == nil {
		err := fmt.Errorf("Invalid address '%s'", opts.IPAddress)
		return err
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
