// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

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
	mech  string
	value []byte
	qual  byte
}
