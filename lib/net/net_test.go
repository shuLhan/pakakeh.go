// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsTypeUDP(t *testing.T) {
	cases := []struct {
		desc string
		netw string
		exp  bool
	}{{
		desc: "Empty network",
	}, {
		desc: "Network is tcp",
		netw: "tcp",
	}, {
		desc: "Network is tcp4",
		netw: "tcp4",
	}, {
		desc: "Network is tcp6",
		netw: "tcp6",
	}, {
		desc: "Network is udp",
		netw: "udp",
		exp:  true,
	}, {
		desc: "Network is udp4",
		netw: "udp4",
		exp:  true,
	}, {
		desc: "Network is udp6",
		netw: "udp6",
		exp:  true,
	}, {
		desc: "Network is ip",
		netw: "ip",
	}, {
		desc: "Network is ip4",
		netw: "ip4",
	}, {
		desc: "Network is ip6",
		netw: "ip6",
	}, {
		desc: "Network is unix",
		netw: "unix",
	}, {
		desc: "Network is unixgram",
		netw: "unixgram",
	}, {
		desc: "Network is unixpacket",
		netw: "unixpacket",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		netType := ConvertStandard(c.netw)

		got := IsTypeUDP(netType)

		test.Assert(t, "IsTypeUDP", c.exp, got, true)
	}
}

func TestIsTypeTCP(t *testing.T) {
	cases := []struct {
		desc string
		netw string
		exp  bool
	}{{
		desc: "Empty network",
	}, {
		desc: "Network is tcp",
		netw: "tcp",
		exp:  true,
	}, {
		desc: "Network is tcp4",
		netw: "tcp4",
		exp:  true,
	}, {
		desc: "Network is tcp6",
		netw: "tcp6",
		exp:  true,
	}, {
		desc: "Network is udp",
		netw: "udp",
	}, {
		desc: "Network is udp4",
		netw: "udp4",
	}, {
		desc: "Network is udp6",
		netw: "udp6",
	}, {
		desc: "Network is ip",
		netw: "ip",
	}, {
		desc: "Network is ip4",
		netw: "ip4",
	}, {
		desc: "Network is ip6",
		netw: "ip6",
	}, {
		desc: "Network is unix",
		netw: "unix",
	}, {
		desc: "Network is unixgram",
		netw: "unixgram",
	}, {
		desc: "Network is unixpacket",
		netw: "unixpacket",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		netType := ConvertStandard(c.netw)

		got := IsTypeTCP(netType)

		test.Assert(t, "IsTypeTCP", c.exp, got, true)
	}
}

func TestIsTypeTransport(t *testing.T) {
	cases := []struct {
		desc string
		netw string
		exp  bool
	}{{
		desc: "Empty network",
	}, {
		desc: "Network is tcp",
		netw: "tcp",
		exp:  true,
	}, {
		desc: "Network is tcp4",
		netw: "tcp4",
		exp:  true,
	}, {
		desc: "Network is tcp6",
		netw: "tcp6",
		exp:  true,
	}, {
		desc: "Network is udp",
		netw: "udp",
		exp:  true,
	}, {
		desc: "Network is udp4",
		netw: "udp4",
		exp:  true,
	}, {
		desc: "Network is udp6",
		netw: "udp6",
		exp:  true,
	}, {
		desc: "Network is ip",
		netw: "ip",
	}, {
		desc: "Network is ip4",
		netw: "ip4",
	}, {
		desc: "Network is ip6",
		netw: "ip6",
	}, {
		desc: "Network is unix",
		netw: "unix",
	}, {
		desc: "Network is unixgram",
		netw: "unixgram",
	}, {
		desc: "Network is unixpacket",
		netw: "unixpacket",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		netType := ConvertStandard(c.netw)

		got := IsTypeTransport(netType)

		test.Assert(t, "IsTypeTransport", c.exp, got, true)
	}
}
