// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"
)

//
// RDataSOA marks the Start Of a zone of Authority RDATA format in resource
// record.
//
// All times are in units of seconds.
//
// Most of these fields are pertinent only for name server maintenance
// operations.  However, MINIMUM is used in all query operations that
// retrieve RRs from a zone.  Whenever a RR is sent in a response to a
// query, the TTL field is set to the maximum of the TTL field from the RR
// and the MINIMUM field in the appropriate SOA.  Thus MINIMUM is a lower
// bound on the TTL field for all RRs in a zone.  Note that this use of
// MINIMUM should occur when the RRs are copied into the response and not
// when the zone is loaded from a master file or via a zone transfer.  The
// reason for this provison is to allow future dynamic update facilities to
// change the SOA RR with known semantics.
//
type RDataSOA struct {
	// The <domain-name> of the name server that was the original or
	// primary source of data for this zone.
	MName string

	// A <domain-name> which specifies the mailbox of the person
	// responsible for this zone.
	RName string

	// The unsigned 32 bit version number of the original copy of the
	// zone.  Zone transfers preserve this value.  This value wraps and
	// should be compared using sequence space arithmetic.
	Serial uint32

	// A 32 bit time interval before the zone should be refreshed.
	Refresh int32

	// A 32 bit time interval that should elapse before a failed refresh
	// should be retried.
	Retry int32

	// A 32 bit time value that specifies the upper limit on the time
	// interval that can elapse before the zone is no longer
	// authoritative.
	Expire int32

	// The unsigned 32 bit minimum TTL field that should be exported with
	// any RR from this zone.
	Minimum uint32
}

//
// String return readable representation of SOA record.
//
func (soa *RDataSOA) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{MName:%s RName:%s Serial:%d Refresh:%d Retry:%d Expire:%d Minimum:%d}",
		soa.MName, soa.RName, soa.Serial, soa.Refresh, soa.Retry,
		soa.Expire, soa.Minimum)

	return b.String()
}
