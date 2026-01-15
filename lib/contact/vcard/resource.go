// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package vcard

// Resource define common resource located in URI or embedded in Data.
type Resource struct {
	Type string
	URI  string
	Data []byte
}
