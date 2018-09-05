// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

//
// RDataText represent generic domain-name or text for NS, CNAME, MB, MG, and
// TEXT RDATA format.
//
type RDataText struct {
	Value []byte
}

// String return string representation of RDATA.
func (text *RDataText) String() string {
	return string(text.Value)
}
