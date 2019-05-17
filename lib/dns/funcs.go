// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"net"
	"strings"

	libnet "github.com/shuLhan/share/lib/net"
)

//
// GetSystemNameServers return list of system name servers by reading
// resolv.conf formatted file in path.
//
// Default path is "/etc/resolv.conf".
//
func GetSystemNameServers(path string) []string {
	if len(path) == 0 {
		path = "/etc/resolv.conf"
	}
	rc, err := libnet.NewResolvConf(path)
	if err != nil {
		return nil
	}
	return rc.NameServers
}

//
// ParseNameServers parse list of nameserver into UDP addresses.
// If one of nameserver is invalid it will stop parsing and return only valid
// nameserver addresses with error.
//
func ParseNameServers(nameservers []string) ([]*net.UDPAddr, error) {
	udpAddrs := make([]*net.UDPAddr, 0)

	for _, ns := range nameservers {
		addr, err := libnet.ParseUDPAddr(ns, DefaultPort)
		if err != nil {
			return udpAddrs, err
		}
		udpAddrs = append(udpAddrs, addr)
	}

	return udpAddrs, nil
}

//
// LookupPTR accept an IP address (either IPv4 or IPv6) and return a single
// answer as domain name on sucess or an error on failed.
// If IP address does not contains PTR record it will return an empty string
// without error.
//
func LookupPTR(client Client, ip net.IP) (answer string, err error) {
	if ip == nil {
		return "", fmt.Errorf("empty IP address")
	}

	revIP, isIPv4 := reverseIP(ip)
	if len(revIP) == 0 {
		return "", fmt.Errorf("invalid IP address %q", ip)
	}

	if isIPv4 {
		revIP = append(revIP, []byte(".in-addr.arpa")...)
	} else {
		revIP = append(revIP, []byte(".ip6.arpa")...)
	}

	msg, err := client.Lookup(true, QueryTypePTR, QueryClassIN, revIP)
	if err != nil {
		return "", err
	}

	rranswers := msg.FilterAnswers(QueryTypePTR)
	if len(rranswers) == 0 {
		return "", nil
	}

	banswer, ok := rranswers[0].RData().([]byte)
	if !ok {
		return "", fmt.Errorf("invalid PTR record data")
	}

	answer = string(banswer)

	return
}

//
// reverseIP reverse the IP address by dot.
//
func reverseIP(ip net.IP) (revIP []byte, isIPv4 bool) {
	strIP := ip.String()

	if strings.Count(strIP, ".") == 3 {
		isIPv4 = true
		revIP = reverseIPv4(strIP)
		return
	}
	if strings.Count(strIP, ":") >= 2 {
		revIP = reverseIPv6(strIP)
		return
	}

	return nil, false
}

//
// reverseIPv4 reverse the IPv4 address. For example, given "127.0.0.1" it
// will return "1.0.0.127".
//
func reverseIPv4(ip string) (rev []byte) {
	addrs := strings.Split(ip, ".")
	for x := len(addrs) - 1; x >= 0; x-- {
		if len(rev) > 0 {
			rev = append(rev, '.')
		}
		rev = append(rev, addrs[x]...)
	}
	return
}

//
// reverseIPv6 reverse the IPv6 address.  For example, given "2001:db8::cb01"
// it will return "1.0.b.c.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.2".
//
func reverseIPv6(ip string) (rev []byte) {
	addrs := strings.Split(ip, ":")

	var notempty int
	for x := 0; x < len(addrs); x++ {
		if len(addrs[x]) != 0 {
			notempty++
		}
	}
	gap := 8 - notempty

	for x := len(addrs) - 1; x >= 0; x-- {
		addr := addrs[x]

		// Fill the gap with "0.0.0.0".
		if len(addr) == 0 {
			for ; gap > 0; gap-- {
				if len(rev) > 0 {
					rev = append(rev, '.')
				}
				rev = append(rev, []byte("0.0.0.0")...)
			}
			continue
		}

		// Reverse the sub address "2001" into "1.0.0.2".
		for y := len(addr) - 1; y >= 0; y-- {
			if len(rev) > 0 {
				rev = append(rev, '.')
			}
			rev = append(rev, addr[y])
		}

		// Fill the sub address with zero.
		for y := len(addr); y < 4; y++ {
			rev = append(rev, []byte(".0")...)
		}
	}

	return
}
