// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"bytes"
	"fmt"
	"strings"
)

// RDataHINFO HINFO records are used to acquire general information about a
// host.  The main use is for protocols such as FTP that can use special
// procedures when talking between machines or operating systems of the same
// type.
type RDataHINFO struct {
	CPU []byte
	OS  []byte
}

// unpack the HINFO RDATA from DNS message.
func (hinfo *RDataHINFO) unpack(packet []byte) error {
	var (
		x    = 0
		size = int(packet[x])
	)
	x++
	hinfo.CPU = bytes.Clone(packet[x : x+size])
	x += size
	size = int(packet[x])
	x++
	hinfo.OS = bytes.Clone(packet[x : x+size])
	return nil
}

// String return readable representation of HINFO record.
func (hinfo *RDataHINFO) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{CPU:%s OS:%s}", hinfo.CPU, hinfo.OS)

	return b.String()
}
