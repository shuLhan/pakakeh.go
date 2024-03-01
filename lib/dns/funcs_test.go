// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestGetSystemNameServers(t *testing.T) {
	type testCase struct {
		path string
		exp  []string
	}

	var (
		cases []testCase
		got   []string
		c     testCase
	)

	cases = append(cases, testCase{
		path: "testdata/resolv.conf",
		exp: []string{
			"127.0.0.1",
		},
	})

	for _, c = range cases {
		t.Log(c.path)

		got = GetSystemNameServers(c.path)

		test.Assert(t, "NameServers", c.exp, got)
	}
}

func TestReverseIP(t *testing.T) {
	type testCase struct {
		ip        string
		exp       []byte
		expIsIPv4 bool
	}

	var (
		ip        net.IP
		cases     []testCase
		c         testCase
		gotIP     []byte
		gotIsIPv4 bool
	)

	cases = []testCase{{
		ip: "",
	}, {
		ip:        "192.0.2.1",
		exp:       []byte("1.2.0.192"),
		expIsIPv4: true,
	}, {
		ip:  "2001:db8::68",
		exp: []byte("8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2"),
	}}

	for _, c = range cases {
		ip = net.ParseIP(c.ip)

		gotIP, gotIsIPv4 = reverseIP(ip)

		test.Assert(t, "reverseIP", c.exp, gotIP)
		test.Assert(t, "isIPv4", c.expIsIPv4, gotIsIPv4)
	}
}

func TestLookupPTR(t *testing.T) {
	type testCase struct {
		exp    string
		expErr string
		ip     net.IP
	}

	var (
		cl    *UDPClient
		cases []testCase
		c     testCase
		got   string
		err   error
	)

	cl, err = NewUDPClient(testServerAddress)
	if err != nil {
		t.Fatal(err)
	}

	cases = []testCase{{
		ip:     nil,
		expErr: "empty IP address",
	}, {
		ip: net.ParseIP("127.0.0.2"),
	}, {
		ip:  net.ParseIP("127.0.0.10"),
		exp: "kilabit.info",
	}, {
		ip:  net.ParseIP("::1"),
		exp: "kilabit.info",
	}, {
		ip:  net.ParseIP("2001:db8::cb01"),
		exp: "kilabit.info",
	}}

	for _, c = range cases {
		t.Logf("ip: %s", c.ip)

		got, err = LookupPTR(cl, c.ip)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "LookupPTR", c.exp, got)
	}
}
