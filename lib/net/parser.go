// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"net"
	"strconv"
)

//
// ParseIPPort parse address into IP and port.
// In case of port is empty or invalid, it will set to default port value in
// defPort.
//
func ParseIPPort(address string, defPort uint16) (ip net.IP, port uint16, err error) {
	var iport int

	shost, sport, err := net.SplitHostPort(address)
	if err != nil {
		shost = address
		err = nil
	}

	ip = net.ParseIP(shost)
	if ip == nil {
		err = ErrHostAddress
		return
	}

	if len(sport) > 0 {
		iport, err = strconv.Atoi(sport)
		if err != nil {
			iport = int(defPort)
			err = nil
		} else {
			if iport < 0 || iport > maxPort {
				iport = int(defPort)
			}
		}
		port = uint16(iport)
	} else {
		port = defPort
	}

	return
}

//
// ParseUDPAddr parse IP address into standard library UDP address.
// In case of port is empty, it will set to default port value in defPort.
//
func ParseUDPAddr(address string, defPort uint16) (udp *net.UDPAddr, err error) {
	ip, port, err := ParseIPPort(address, defPort)
	if err != nil {
		return
	}

	udp = &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}

	return
}

//
// ParseTCPAddr parse IP address into standard library TCP address.
// In case of port is empty, it will set to default port value in defPort.
//
func ParseTCPAddr(address string, defPort uint16) (udp *net.TCPAddr, err error) {
	ip, port, err := ParseIPPort(address, defPort)
	if err != nil {
		return
	}

	udp = &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}

	return
}
