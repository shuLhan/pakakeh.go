// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package net provide constants and library for networking.
package net

import (
	"errors"
	"strings"
)

const (
	maxPort = (1 << 16) - 1
)

// List of error messages.
var (
	ErrHostAddress = errors.New("invalid host address")
)

// Type of network.
type Type uint16

// List of possible network type.
const (
	TypeInvalid Type = 0
	TypeTCP     Type = 1 << iota
	TypeTCP4
	TypeTCP6
	TypeUDP
	TypeUDP4
	TypeUDP6
	TypeIP
	TypeIP4
	TypeIP6
	TypeUnix
	TypeUnixGram
	TypeUnixPacket
)

//
// ConvertStandard library network value from string to Type.
// It will return TypeInvalid (0) if network is unknown.
//
func ConvertStandard(network string) Type {
	network = strings.ToLower(network)

	switch network {
	case "tcp":
		return TypeTCP
	case "tcp4":
		return TypeTCP4
	case "tcp6":
		return TypeTCP6
	case "udp":
		return TypeUDP
	case "udp4":
		return TypeUDP4
	case "udp6":
		return TypeUDP6
	case "ip":
		return TypeIP
	case "ip4":
		return TypeIP4
	case "ip6":
		return TypeIP6
	case "unix":
		return TypeUnix
	case "unixgram":
		return TypeUnixGram
	case "unixpacket":
		return TypeUnixPacket
	}
	return TypeInvalid
}

//
// IsTypeTCP will return true if t is type of TCP(4,6); otherwise it will
// return false.
//
func IsTypeTCP(t Type) bool {
	if t == TypeTCP || t == TypeTCP4 || t == TypeTCP6 {
		return true
	}
	return false
}

//
// IsTypeUDP will return true if t is type of UDP(4,6); otherwise it will
// return false.
//
func IsTypeUDP(t Type) bool {
	if t == TypeUDP || t == TypeUDP4 || t == TypeUDP6 {
		return true
	}
	return false
}

//
// IsTypeTransport will return true if t is type of transport layer, i.e.
// tcp(4,6) or udp(4,6); otherwise it will return false.
//
func IsTypeTransport(t Type) bool {
	return IsTypeTCP(t) || IsTypeUDP(t)
}
