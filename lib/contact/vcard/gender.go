// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package vcard

// Gender contains contact's sex and description.
//
// Sex may contain one of this value: M (male), F (female), O (other), N (none),
// or U (unknown).
type Gender struct {
	Desc string
	Sex  rune
}
