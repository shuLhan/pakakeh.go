// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package vcard

// Relation define a contact relation to other contact URI.
type Relation struct {
	Type string
	URI  string
}
