// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestServerOptionsInit(t *testing.T) {
	type testCase struct {
		desc     string
		so       *ServerOptions
		exp      *ServerOptions
		expError string
	}

	var (
		ip     = net.ParseIP("0.0.0.0")
		defSoa = NewRDataSOA(``, ``)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc: "With empty value",
		so:   &ServerOptions{},
		exp: &ServerOptions{
			ListenAddress:   "0.0.0.0:53",
			HTTPIdleTimeout: defaultHTTPIdleTimeout,
			SOA:             *defSoa,
			PruneDelay:      time.Hour,
			PruneThreshold:  -1 * time.Hour,
			ip:              ip,
			port:            53,
		},
	}, {
		desc: "With invalid IP address",
		so: &ServerOptions{
			ListenAddress: "0.0.0",
		},
		expError: `dns: invalid IP address '0.0.0'`,
	}, {
		desc: "With no valid name servers",
		so: &ServerOptions{
			NameServers: []string{
				"udp://localhost",
			},
		},
		expError: `dns: no valid name servers`,
	}, {
		desc: "With valid name servers",
		so: &ServerOptions{
			NameServers: []string{
				"udp://127.0.0.1",
			},
		},
		exp: &ServerOptions{
			ListenAddress:   "0.0.0.0:53",
			HTTPIdleTimeout: defaultHTTPIdleTimeout,
			NameServers: []string{
				"udp://127.0.0.1",
			},
			SOA:            *defSoa,
			PruneDelay:     time.Hour,
			PruneThreshold: -1 * time.Hour,
			ip:             ip,
			port:           53,
			primaryUDP: []net.Addr{
				&net.UDPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 53,
				},
			},
			primaryTCP: []net.Addr{
				&net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 53,
				},
			},
		},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		err = c.so.init()
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		test.Assert(t, "ServerOptions", c.exp, c.so)
	}
}

func TestServerOptionsParseNameServers(t *testing.T) {
	type testCase struct {
		desc          string
		nameServers   []string
		expUDPServers []net.Addr
		expTCPServers []net.Addr
		expDoHServers []string
	}

	var (
		so = &ServerOptions{}
		ip = net.ParseIP("127.0.0.1")

		cases []testCase
		c     testCase
	)

	cases = []testCase{{
		desc: "With empty input",
	}, {
		desc: "With invalid URI",
		nameServers: []string{
			"://127.0.0.1",
		},
	}, {
		desc: "With valid hostname on UDP",
		nameServers: []string{
			"udp://localhost:53",
		},
	}, {
		desc: "With valid hostname on TCP",
		nameServers: []string{
			"tcp://localhost:53",
		},
	}, {
		desc: "With no scheme",
		nameServers: []string{
			"127.0.0.1",
		},
		expUDPServers: []net.Addr{&net.UDPAddr{
			IP:   ip,
			Port: 53,
		}},
		expTCPServers: []net.Addr{&net.TCPAddr{
			IP:   ip,
			Port: 53,
		}},
	}, {
		desc: "With valid name servers",
		nameServers: []string{
			"udp://127.0.0.1",
			"tcp://127.0.0.1:5353",
			"https://localhost/dns-query",
		},
		expUDPServers: []net.Addr{&net.UDPAddr{
			IP:   ip,
			Port: 53,
		}},
		expTCPServers: []net.Addr{
			&net.TCPAddr{
				IP:   ip,
				Port: 53,
			},
			&net.TCPAddr{
				IP:   ip,
				Port: 5353,
			},
		},
		expDoHServers: []string{
			"https://localhost/dns-query",
		},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		so.NameServers = c.nameServers

		so.initNameServers()

		test.Assert(t, "primaryUDP", c.expUDPServers, so.primaryUDP)
		test.Assert(t, "primaryTCP", c.expTCPServers, so.primaryTCP)
		test.Assert(t, "primaryDoh", c.expDoHServers, so.primaryDoh)
	}
}
