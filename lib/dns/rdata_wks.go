// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package dns

import (
	"fmt"
	"strings"
)

// RDataWKS The WKS record is used to describe the well known services
// supported by a particular protocol on a particular internet address.  The
// PROTOCOL field specifies an IP protocol number, and the bit map has one bit
// per port of the specified protocol.  The first bit corresponds to port 0,
// the second to port 1, etc.  If the bit map does not include a bit for a
// protocol of interest, that bit is assumed zero.  The appropriate values and
// mnemonics for ports and protocols are specified in [RFC-1010].
//
// For example, if PROTOCOL=TCP (6), the 26th bit corresponds to TCP port
// 25 (SMTP).  If this bit is set, a SMTP server should be listening on TCP
// port 25; if zero, SMTP service is not supported on the specified
// address.
//
// The purpose of WKS RRs is to provide availability information for
// servers for TCP and UDP.  If a server supports both TCP and UDP, or has
// multiple Internet addresses, then multiple WKS RRs are used.
type RDataWKS struct {
	Address  []byte
	BitMap   []byte
	Protocol byte
}

// unpack the WKS record from DNS RR in packet.
func (wks *RDataWKS) unpack(packet []byte) error {
	wks.Address = append(wks.Address, packet[0:4]...)
	wks.Protocol = packet[4]
	wks.BitMap = packet[5:]
	return nil
}

// String return readable representation of WKS record.
func (wks *RDataWKS) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{Address:%s Protocol:%d BitMap:%s}", wks.Address,
		wks.Protocol, wks.BitMap)

	return b.String()
}
