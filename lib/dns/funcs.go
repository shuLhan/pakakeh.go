// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"errors"
	"fmt"
	"net"

	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

// GetSystemNameServers return list of system name servers by reading
// resolv.conf formatted file in path.
//
// Default path is "/etc/resolv.conf".
func GetSystemNameServers(path string) []string {
	var (
		rc  *libnet.ResolvConf
		err error
	)

	if len(path) == 0 {
		path = "/etc/resolv.conf"
	}
	rc, err = libnet.NewResolvConf(path)
	if err != nil {
		return nil
	}
	return rc.NameServers
}

// ParseNameServers parse list of nameserver into UDP addresses.
// If one of nameserver is invalid it will stop parsing and return only valid
// nameserver addresses with error.
func ParseNameServers(nameservers []string) (udpAddrs []*net.UDPAddr, err error) {
	var (
		ns   string
		addr *net.UDPAddr
	)

	for _, ns = range nameservers {
		addr, err = libnet.ParseUDPAddr(ns, DefaultPort)
		if err != nil {
			return udpAddrs, err
		}
		udpAddrs = append(udpAddrs, addr)
	}

	return udpAddrs, nil
}

// LookupPTR accept an IP address (either IPv4 or IPv6) and return a single
// answer as domain name on success or an error on failed.
// If IP address does not contains PTR record it will return an empty string
// without error.
func LookupPTR(client Client, ip net.IP) (answer string, err error) {
	if ip == nil {
		return ``, errors.New(`empty IP address`)
	}

	var (
		msg       *Message
		q         MessageQuestion
		rranswers []ResourceRecord
		revIP     []byte
		isIPv4    bool
	)

	revIP, isIPv4 = reverseIP(ip)
	if len(revIP) == 0 {
		return "", fmt.Errorf("invalid IP address %q", ip)
	}

	if isIPv4 {
		revIP = append(revIP, []byte(".in-addr.arpa")...)
	} else {
		revIP = append(revIP, []byte(".ip6.arpa")...)
	}

	q.Name = string(revIP)
	q.Type = RecordTypePTR

	msg, err = client.Lookup(q, true)
	if err != nil {
		return "", err
	}

	rranswers = msg.FilterAnswers(RecordTypePTR)
	if len(rranswers) == 0 {
		return "", nil
	}

	var ok bool
	answer, ok = rranswers[0].Value.(string)
	if !ok {
		return ``, errors.New(`invalid PTR record data`)
	}

	return answer, nil
}

// reverseIP reverse the IP address by dot.
func reverseIP(ip net.IP) (revIP []byte, isIPv4 bool) {
	isIPv4 = libnet.IsIPv4(ip)
	if isIPv4 {
		revIP = reverseByDot([]byte(ip.String()))
		return
	}
	if libnet.IsIPv6(ip) {
		revIP = reverseByDot(libnet.ToDotIPv6(ip))
		return
	}
	return nil, false
}

// reverseByDot reverse the IP address by dot.
// For example, IPv4 with address "127.0.0.1" it will return "1.0.0.127".
// For IPv6 with address "2001:db8::cb01" it will return
// "1.0.b.c.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.2".
func reverseByDot(ip []byte) (rev []byte) {
	var (
		addrs = bytes.Split(ip, []byte{'.'})

		x int
	)
	for x = len(addrs) - 1; x >= 0; x-- {
		if len(rev) > 0 {
			rev = append(rev, '.')
		}
		rev = append(rev, addrs[x]...)
	}
	return
}
