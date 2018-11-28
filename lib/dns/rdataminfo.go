// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"
)

//
// RDataMINFO define a resource record for type MINFO.
//
type RDataMINFO struct {
	// A <domain-name> which specifies a mailbox which is responsible for
	// the mailing list or mailbox.  If this domain name names the root,
	// the owner of the MINFO RR is responsible for itself.  Note that
	// many existing mailing lists use a mailbox X-request for the RMAILBX
	// field of mailing list X, e.g., Msgroup-request for Msgroup.  This
	// field provides a more general mechanism.
	RMailBox []byte

	// A <domain-name> which specifies a mailbox which is to receive error
	// messages related to the mailing list or mailbox specified by the
	// owner of the MINFO RR (similar to the ERRORS-TO: field which has
	// been proposed).  If this domain name names the root, errors should
	// be returned to the sender of the message.
	EmailBox []byte
}

//
// String return readable representation of MINFO record.
//
func (minfo *RDataMINFO) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{RMailBox:%s EmailBox:%s}", minfo.RMailBox,
		minfo.EmailBox)

	return b.String()
}
