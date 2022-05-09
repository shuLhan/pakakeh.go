// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"
)

// RDataOPT define format of RDATA for OPT.
//
// The extended RCODE and flags, which OPT stores in the RR Time to Live
// (TTL) field, contains ExtRCode, Version
type RDataOPT struct {
	// Varies per OPTION-CODE.  MUST be treated as a bit field.
	Data []byte

	// Assigned by the Expert Review process as defined by the DNSEXT
	// working group and the IESG.
	Code uint16

	// Size (in octets) of OPTION-DATA.
	Length uint16

	// Forms the upper 8 bits of extended 12-bit RCODE (together with the
	// 4 bits defined in [RFC1035].  Note that EXTENDED-RCODE value 0
	// indicates that an unextended RCODE is in use (values 0 through 15).
	ExtRCode byte

	// Indicates the implementation level of the setter.  Full conformance
	// with this specification is indicated by version '0'.  Requestors
	// are encouraged to set this to the lowest implemented level capable
	// of expressing a transaction, to minimise the responder and network
	// load of discovering the greatest common implementation level
	// between requestor and responder.  A requestor's version numbering
	// strategy MAY ideally be a run-time configuration option.
	Version byte

	// DNSSEC OK bit as defined by [RFC3225].
	DO bool
}

// String return readable representation of OPT record.
func (opt *RDataOPT) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{ExtRCode:%d Version:%d DO:%v Code:%d Length:%d Data:%s}",
		opt.ExtRCode, opt.Version, opt.DO, opt.Code, opt.Length,
		opt.Data)

	return b.String()
}
