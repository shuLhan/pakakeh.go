// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"
)

//
// RDataHINFO HINFO records are used to acquire general information about a
// host.  The main use is for protocols such as FTP that can use special
// procedures when talking between machines or operating systems of the same
// type.
//
type RDataHINFO struct {
	CPU []byte
	OS  []byte
}

//
// unpack the HINFO RDATA from DNS message.
//
func (hinfo *RDataHINFO) unpack(packet []byte) error {
	x := 0
	for ; x < len(packet); x++ {
		if packet[x] == 0 {
			break
		}
	}
	hinfo.CPU = append(hinfo.CPU, packet[0:x]...)
	x++
	hinfo.OS = append(hinfo.OS, packet[x:]...)

	return nil
}

//
// String return readable representation of HINFO record.
//
func (hinfo *RDataHINFO) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{CPU:%s OS:%s}", hinfo.CPU, hinfo.OS)

	return b.String()
}
