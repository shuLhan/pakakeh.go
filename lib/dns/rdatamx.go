// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"

	libnet "github.com/shuLhan/share/lib/net"
)

const (
	defMXPreference int16 = 10
)

// RDataMX MX records cause type A additional section processing for the host
// specified by EXCHANGE.  The use of MX RRs is explained in detail in
// [RFC-974].
type RDataMX struct {
	// A 16 bit integer which specifies the preference given to this RR
	// among others at the same owner.  Lower values are preferred.
	Preference int16

	// A <domain-name> which specifies a host willing to act as a mail
	// exchange for the owner name.
	Exchange string
}

//
// String return readable representation of MX record.
//
func (mx *RDataMX) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{Preference:%d Exchange:%s}", mx.Preference,
		mx.Exchange)

	return b.String()
}

//
// initAndValidate initialize and validate the MX fields.
//
func (mx *RDataMX) initAndValidate() error {
	if mx.Preference <= 0 {
		mx.Preference = defMXPreference
	}
	if !libnet.IsHostnameValid([]byte(mx.Exchange), true) {
		return fmt.Errorf("invalid or empty MX Exchange %q",
			string(mx.Exchange))
	}
	return nil
}
