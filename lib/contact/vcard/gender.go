// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcard

// Gender contains contact's sex and description.
//
// Sex may contain one of this value: M (male), F (female), O (other), N (none),
// or U (unknown).
type Gender struct {
	Desc string
	Sex  rune
}
