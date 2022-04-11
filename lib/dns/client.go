// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

//
// Client is interface that implement sending and receiving DNS message.
//
type Client interface {
	Close() error
	Lookup(q MessageQuestion, allowRecursion bool) (*Message, error)
	Query(req *Message) (*Message, error)
	RemoteAddr() string
	SetRemoteAddr(addr string) error
	SetTimeout(t time.Duration)
}

//
// NewClient create new DNS client using the name server URL.
// The name server URL is defined in the same format as
// ServerOptions.NameServers.
// The isInsecure parameter only usable for DNS over TLS (DoT) and DNS over
// HTTPS (DoH).
//
// For example,
//
//  - "udp://127.0.0.1:53" for UDP client.
//  - "tcp://127.0.0.1:53" for TCP client.
//  - "https://127.0.0.1:853" (HTTPS with IP address) for DoT.
//  - "https://localhost/dns-query" (HTTPS with domain name) for DoH.
//
func NewClient(nsUrl string, isInsecure bool) (cl Client, err error) {
	var (
		logp = "NewClient"

		urlNS  *url.URL
		ip     net.IP
		iphost string
		port   string
		ipport []string
	)

	urlNS, err = url.Parse(nsUrl)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid name server URL: %q", logp, nsUrl)
	}

	ipport = strings.Split(urlNS.Host, ":")
	switch len(ipport) {
	case 1:
		iphost = ipport[0]
	case 2:
		iphost = ipport[0]
		port = ipport[1]
	default:
		return nil, fmt.Errorf("%s: invalid name server URL: %q", logp, nsUrl)
	}

	switch urlNS.Scheme {
	case "udp":
		cl, err = NewUDPClient(urlNS.Host)
	case "tcp":
		cl, err = NewTCPClient(urlNS.Host)
	case "https":
		ip = net.ParseIP(iphost)
		if ip == nil {
			cl, err = NewDoHClient(nsUrl, isInsecure)
		} else {
			if len(port) == 0 {
				port = "853"
			}
			cl, err = NewDoTClient(iphost+":"+port, isInsecure)
		}
	default:
		return nil, fmt.Errorf("%s: unknown scheme %q", logp, urlNS.Scheme)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return cl, nil
}
