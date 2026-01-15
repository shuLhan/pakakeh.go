// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package net

import (
	"net"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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

		test.Assert(t, "IsTypeUDP", c.exp, got)
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

		test.Assert(t, "IsTypeTCP", c.exp, got)
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

		test.Assert(t, "IsTypeTransport", c.exp, got)
	}
}

func TestToDotIPv6(t *testing.T) {
	cases := []struct {
		ip  net.IP
		exp []byte
	}{{
		ip:  net.ParseIP("2001:db8::68"),
		exp: []byte("2.0.0.1.0.d.b.8.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.6.8"),
	}, {
		ip:  net.ParseIP("::1"),
		exp: []byte("0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1"),
	}}

	for _, c := range cases {
		got := ToDotIPv6(c.ip)
		test.Assert(t, "ToDotIPv6", c.exp, got)
	}
}

func TestWaitAlive_tcp(t *testing.T) {
	var (
		address = `127.0.0.1:27392`
		tcpAddr = net.TCPAddr{
			IP:   net.ParseIP(`127.0.0.1`),
			Port: 27392,
		}
		errq = make(chan error, 1)

		tcpListener *net.TCPListener
	)

	go func() {
		var err2 error
		tcpListener, err2 = net.ListenTCP(`tcp`, &tcpAddr)
		errq <- err2
	}()

	var err = WaitAlive(`tcp`, address, 5*time.Second)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}
	err = <-errq
	if err != nil {
		t.Fatal(err)
	}
	_ = tcpListener.Close()
}

func TestWaitAlive_udp(t *testing.T) {
	var (
		address = `127.0.0.1:27392`
		udpAddr = net.UDPAddr{
			IP:   net.ParseIP(`127.0.0.1`),
			Port: 27392,
		}
		errq = make(chan error, 1)

		udpConn *net.UDPConn
	)

	go func() {
		var err2 error
		udpConn, err2 = net.ListenUDP(`udp`, &udpAddr)
		errq <- err2
	}()

	var err = WaitAlive(`udp`, address, 5*time.Second)
	if err != nil {
		t.Fatalf(`want no error, got %s`, err)
	}
	err = <-errq
	if err != nil {
		t.Fatal(err)
	}
	_ = udpConn.Close()
}
