// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package net provide constants and library for networking.
package net

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

const (
	maxPort = (1 << 16) - 1
)

// ErrHostAddress define an error if address of connection is invalid.
var ErrHostAddress = errors.New("invalid host address")

// ErrReadTimeout define an error when [Read] operation receive no data
// after waiting for specific duration.
var ErrReadTimeout = errors.New(`read timeout`)

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

// ConvertStandard library network value from string to Type.
// It will return TypeInvalid (0) if network is unknown.
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

// IsTypeTCP will return true if t is type of TCP(4,6); otherwise it will
// return false.
func IsTypeTCP(t Type) bool {
	if t == TypeTCP || t == TypeTCP4 || t == TypeTCP6 {
		return true
	}
	return false
}

// IsTypeUDP will return true if t is type of UDP(4,6); otherwise it will
// return false.
func IsTypeUDP(t Type) bool {
	if t == TypeUDP || t == TypeUDP4 || t == TypeUDP6 {
		return true
	}
	return false
}

// IsTypeTransport will return true if t is type of transport layer, i.e.
// tcp(4,6) or udp(4,6); otherwise it will return false.
func IsTypeTransport(t Type) bool {
	return IsTypeTCP(t) || IsTypeUDP(t)
}

// Read packet from network.
//
// If the conn parameter is nil it will return [net.ErrClosed].
//
// The bufsize parameter set the size of buffer for each read operation,
// default to 1024 if not set or invalid (less than 0 or greater than
// 65535).
//
// The timeout parameter set how long to wait for data before considering
// it as failed.
// If its not set, less or equal to 0, it will wait forever.
// If no data received and timeout is set, it will return [ErrReadTimeout].
//
// If there is data received and connection closed at the same time, it will
// return the data first without error.
// The subsequent Read will return empty packet with [ErrClosed].
func Read(conn net.Conn, bufsize int, timeout time.Duration) (packet []byte, err error) {
	var logp = `Read`

	if conn == nil {
		return nil, fmt.Errorf(`%s: %w`, logp, net.ErrClosed)
	}
	if bufsize <= 0 || bufsize > math.MaxUint16 {
		bufsize = 1024
	}

	var (
		buf = make([]byte, bufsize)
		n   int
	)

	if timeout > 0 {
		err = conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}
	for {
		n, err = conn.Read(buf)
		if err != nil {
			var neterr net.Error
			if errors.As(err, &neterr) && neterr.Timeout() {
				return nil, ErrReadTimeout
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		if n == 0 {
			// Connection closed by peer.
			break
		}

		packet = append(packet, buf[:n]...)
		if n < len(buf) {
			break
		}
		// Keep reading if we read full buffer.
	}
	if len(packet) == 0 {
		return nil, net.ErrClosed
	}
	return packet, nil

}

// ToDotIPv6 convert the IPv6 address format from "::1" format into
// "0.0.0.0 ... 0.0.0.1".
//
// This function only useful for expanding SPF macro "i" or when generating
// query for DNS PTR.
func ToDotIPv6(ip net.IP) (out []byte) {
	addrs := strings.Split(ip.String(), ":")

	var notempty int
	for x := 0; x < len(addrs); x++ {
		if len(addrs[x]) != 0 {
			notempty++
		}
	}
	gap := 8 - notempty

	for x := 0; x < len(addrs); x++ {
		addr := addrs[x]

		// Fill the gap "::" with one or more "0.0.0.0".
		if len(addr) == 0 {
			for ; gap > 0; gap-- {
				if len(out) > 0 {
					out = append(out, '.')
				}
				out = append(out, []byte("0.0.0.0")...)
			}
			continue
		}

		// Fill the sub address with zero.
		for y := len(addr); y < 4; y++ {
			if len(out) > 0 {
				out = append(out, '.')
			}
			out = append(out, '0')
		}

		for y := 0; y < len(addr); y++ {
			if len(out) > 0 {
				out = append(out, '.')
			}
			out = append(out, addr[y])
		}
	}

	return out
}

// WaitAlive try to connect to network at address until timeout reached.
// If connection cannot established it will return an error.
//
// Unlike [net.DialTimeout], this function will retry not returning an error
// immediately if the address has not ready yet.
func WaitAlive(network, address string, timeout time.Duration) (err error) {
	var (
		logp        = `WaitAlive`
		dialTimeout = 100 * time.Millisecond
		total       = dialTimeout
		dialer      = net.Dialer{Timeout: timeout}
		ticker      = time.NewTicker(dialTimeout)

		conn net.Conn
	)

	for total < timeout {
		<-ticker.C
		conn, err = dialer.Dial(network, address)
		if err != nil {
			total += dialTimeout
			continue
		}
		// Connection successfully established.
		ticker.Stop()
		_ = conn.Close()
		return nil
	}
	ticker.Stop()
	return fmt.Errorf(`%s: timeout connecting to %s after %s`, logp, address, timeout)
}
