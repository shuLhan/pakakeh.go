// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

// RDataOPTVar contains the option in OPT RDATA.
type RDataOPTVar struct {
	// Varies per Code.
	// MUST be treated as a bit field.
	Data []byte

	// Assigned by the Expert Review process as defined by the DNSEXT
	// working group and the IESG.
	Code uint16
}
