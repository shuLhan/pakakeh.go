// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"net"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseIPPort(t *testing.T) {
	localIP := net.ParseIP("127.0.0.1")

	cases := []struct {
		desc        string
		address     string
		expHostname string
		expIP       net.IP
		expPort     uint16
	}{{
		desc:    "Empty address",
		expPort: 1,
	}, {
		desc:        "With hostname",
		address:     "address",
		expHostname: "address",
		expPort:     1,
	}, {
		desc:        "With hostname and port",
		address:     "address:555",
		expHostname: "address",
		expPort:     555,
	}, {
		desc:        "With empty port",
		address:     "127.0.0.1",
		expHostname: "127.0.0.1",
		expIP:       localIP,
		expPort:     1,
	}, {
		desc:        "With invalid port",
		address:     "127.0.0.1:a",
		expHostname: "127.0.0.1",
		expIP:       localIP,
		expPort:     1,
	}, {
		desc:        "With invalid port < 0",
		address:     "127.0.0.1:-1",
		expHostname: "127.0.0.1",
		expIP:       localIP,
		expPort:     1,
	}, {
		desc:        "With invalid port > 65535",
		address:     "127.0.0.1:65536",
		expHostname: "127.0.0.1",
		expIP:       localIP,
		expPort:     1,
	}, {
		desc:        "With valid port",
		address:     "127.0.0.1:555",
		expHostname: "127.0.0.1",
		expIP:       localIP,
		expPort:     555,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		hostname, ip, port := ParseIPPort(c.address, 1)

		test.Assert(t, "hostname", c.expHostname, hostname, true)
		test.Assert(t, "ip", c.expIP, ip, true)
		test.Assert(t, "port", c.expPort, port, true)
	}
}
