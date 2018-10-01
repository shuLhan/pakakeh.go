// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcard

//
// Relation define a contact relation to other contact URI.
//
type Relation struct {
	Type string
	URI  string
}
