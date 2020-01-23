// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"fmt"
	"net"
	"strconv"
)

//
// ParseIPPort parse address into IP and port.
// If address is not an IP address, it will return the address as hostname
// (without port number if its exist) and nil on ip.
// In case of port is empty or invalid, it will set to defPort.
//
func ParseIPPort(address string, defPort uint16) (hostname string, ip net.IP, port uint16) {
	var iport int

	shost, sport, err := net.SplitHostPort(address)
	if err != nil {
		shost = address
	}

	ip = net.ParseIP(shost)
	if ip == nil {
		hostname = shost
	}
	if len(sport) > 0 {
		iport, err = strconv.Atoi(sport)
		if err != nil {
			iport = int(defPort)
		} else if iport < 0 || iport > maxPort {
			iport = int(defPort)
		}
		port = uint16(iport)
	} else {
		port = defPort
	}

	return hostname, ip, port
}

//
// ParseUDPAddr parse IP address into standard library UDP address.
// If address is not contains IP address, it will return nil with error.
// In case of port is empty, it will set to default port value in defPort.
//
func ParseUDPAddr(address string, defPort uint16) (udp *net.UDPAddr, err error) {
	_, ip, port := ParseIPPort(address, defPort)
	if ip == nil {
		return nil, fmt.Errorf("net: invalid IP address: %s", address)
	}

	udp = &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}

	return
}

//
// ParseTCPAddr parse IP address into standard library TCP address.
// If address is not contains IP address, it will return nil with error.
// In case of port is empty, it will set to default port value in defPort.
//
func ParseTCPAddr(address string, defPort uint16) (udp *net.TCPAddr, err error) {
	_, ip, port := ParseIPPort(address, defPort)
	if ip == nil {
		return nil, fmt.Errorf("net: invalid IP address: %s", address)
	}

	udp = &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}

	return
}
