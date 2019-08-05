// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

// List of known qualifier for directive.
const (
	qualifierPass     byte = '+'
	qualifierFail     byte = '-'
	qualifierSoftfail byte = '~'
	qualifierNeutral  byte = '?'
)

// List of mechanism for directive.
const (
	mechanismAll     = "all"
	mechanismInclude = "include"
	mechanismA       = "a"
	mechanismMx      = "mx"
	mechanismPtr     = "ptr"
	mechanismIP4     = "ip4"
	mechanismIP6     = "ip6"
	mechanismExist   = "exist"
)

type directive struct {
	qual  byte
	mech  string
	value []byte
}
