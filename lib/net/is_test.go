package net

import (
	"net"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsHostnameValid(t *testing.T) {
	cases := []struct {
		in     []byte
		isFQDN bool
		exp    bool
	}{{
		in: []byte(""),
	}, {
		in: []byte("-1a"),
	}, {
		in: []byte(".1a."),
	}, {
		in: []byte("1a-"),
	}, {
		in:  []byte("a"),
		exp: true,
	}, {
		in:  []byte("_a"),
		exp: true,
	}, {
		in:  []byte("11"),
		exp: true,
	}, {
		in:  []byte("a1"),
		exp: true,
	}, {
		in:  []byte("a-1"),
		exp: true,
	}, {
		in:  []byte("a.1"),
		exp: true,
	}, {
		in:     []byte("a"),
		isFQDN: true,
		exp:    false,
	}, {
		in:     []byte("a.b"),
		isFQDN: true,
		exp:    true,
	}}

	for _, c := range cases {
		t.Logf("input: %s", c.in)

		got := IsHostnameValid(c.in, c.isFQDN)

		test.Assert(t, "IsHostnameValid", c.exp, got)
	}
}

func TestIsIPv4(t *testing.T) {
	cases := []struct {
		ip  string
		exp bool
	}{{
		ip: "",
	}, {
		ip: "127.0.0",
	}, {
		ip:  "127.0.0.1",
		exp: true,
	}, {
		ip: "127.0.0.1.",
	}}

	for _, c := range cases {
		ip := net.ParseIP(c.ip)
		got := IsIPv4(ip)
		test.Assert(t, "IsIPv4: "+c.ip, c.exp, got)
	}
}

func TestIsIPv6(t *testing.T) {
	cases := []struct {
		ip  string
		exp bool
	}{{
		ip: "",
	}, {
		ip: "127.0.0.1:50",
	}, {
		ip: ":1",
	}, {
		ip:  "::1",
		exp: true,
	}}

	for _, c := range cases {
		ip := net.ParseIP(c.ip)
		got := IsIPv6(ip)
		test.Assert(t, "IsIPv4", c.exp, got)
	}
}
