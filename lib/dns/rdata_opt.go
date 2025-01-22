// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// RDataOPT define format of RDATA for OPT.
//
// The extended RCODE and flags, which OPT stores in the RR Time to Live
// (TTL) field, contains ExtRCode, Version, and DNSSEC OK flag.
type RDataOPT struct {
	// ListVar list of pair of code-value inside the RDATA.
	ListVar []RDataOPTVar

	// Forms the upper 8 bits of extended 12-bit RCODE (together with
	// the 4 bits defined message header).
	// Note that the value of 0 indicates that the RCODE in message
	// header is in use (values 0 through 15).
	ExtRCode byte

	// Indicates the implementation level of the setter.
	// Full conformance with this specification is indicated by version
	// '0'.
	// Requestors are encouraged to set this to the lowest implemented
	// level capable of expressing a transaction, to minimise the
	// responder and network load of discovering the greatest common
	// implementation level between requestor and responder.
	// A requestor's version numbering strategy MAY ideally be a
	// run-time configuration option.
	Version byte

	// DNSSEC OK bit as defined by [RFC3225].
	DO bool
}

// String return readable representation of OPT record.
func (opt *RDataOPT) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{ExtRCode:%d Version:%d DO:%v}",
		opt.ExtRCode, opt.Version, opt.DO)

	return b.String()
}

// pack the ListVar for RDATA.
func (opt *RDataOPT) pack() (rdata []byte) {
	var optvar RDataOPTVar
	for _, optvar = range opt.ListVar {
		rdata = binary.BigEndian.AppendUint16(rdata, optvar.Code)
		rdata = binary.BigEndian.AppendUint16(rdata, uint16(len(optvar.Data)))
		rdata = append(rdata, optvar.Data...)
	}
	return rdata
}

// unpack extended-rcode with flags from ext (RR TTL), and RDATA from raw
// packet.
func (opt *RDataOPT) unpack(rdlen int, packet []byte) (err error) {
	var x int

	for x < rdlen {
		var optvar = RDataOPTVar{}

		optvar.Code = binary.BigEndian.Uint16(packet[x:])
		x += 2

		var optlen = int(binary.BigEndian.Uint16(packet[x:]))
		x += 2

		if x+optlen > len(packet) {
			return fmt.Errorf(`option-length is out of range (want=%d, got=%d)`, optlen, len(packet))
		}

		optvar.Data = append(optvar.Data, packet[x:x+optlen]...)
		x += optlen

		opt.ListVar = append(opt.ListVar, optvar)
	}
	return nil
}
