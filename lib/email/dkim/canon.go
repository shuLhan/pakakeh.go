// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// Canon define type of canonicalization algorithm.
//
type Canon byte

//
// List of valid and known canonicalization algorithms.
//
const (
	CanonSimple Canon = iota // "simple" (default)
	CanonRelaxed
)

//
// canonNames contains mapping between canonical type and their human
// readabale names.
//
var canonNames = map[Canon][]byte{
	CanonSimple:  []byte("simple"),
	CanonRelaxed: []byte("relaxed"),
}
