// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

// RDataMX MX records cause type A additional section processing for the host
// specified by EXCHANGE.  The use of MX RRs is explained in detail in
// [RFC-974].
type RDataMX struct {
	// A 16 bit integer which specifies the preference given to this RR
	// among others at the same owner.  Lower values are preferred.
	Preference int16

	// A <domain-name> which specifies a host willing to act as a mail
	// exchange for the owner name.
	Exchange []byte
}
