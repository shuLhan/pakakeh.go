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
		desc    string
		address string
		expErr  error
		expIP   net.IP
		expPort uint16
	}{{
		desc:   "Empty address",
		expErr: ErrHostAddress,
	}, {
		desc:    "Invalid address",
		address: "address",
		expErr:  ErrHostAddress,
	}, {
		desc:    "Empty port",
		address: "127.0.0.1",
		expIP:   localIP,
		expPort: 1,
	}, {
		desc:    "Invalid port",
		address: "127.0.0.1:a",
		expIP:   localIP,
		expPort: 1,
	}, {
		desc:    "Invalid port < 0",
		address: "127.0.0.1:-1",
		expIP:   localIP,
		expPort: 1,
	}, {
		desc:    "Invalid port > 65535",
		address: "127.0.0.1:65536",
		expIP:   localIP,
		expPort: 1,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ip, port, err := ParseIPPort(c.address, 1)
		if err != nil {
			test.Assert(t, "error", c.expErr, err, true)
			continue
		}

		test.Assert(t, "ip", c.expIP, ip, true)
		test.Assert(t, "port", c.expPort, port, true)
	}
}
