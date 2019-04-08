// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"net"

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
