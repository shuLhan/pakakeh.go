// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

// List of known qualifier for directive.
const (
	qualifierPass     int = iota // "+"
	qualifierFail                // "-"
	qualifierSoftfail            // "~"
	qualifierNeutral             // "?"
)

// List of mechanism for directive.
const (
	mechanismAll int = iota
	mechanismInclude
	mechanismA
	mechanismMx
	mechanismPtr
	mechanismIp4
	mechanismIp6
	mechanismExists
)

type directive struct {
	qual  int
	mech  int
	value []byte
}
