// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Phone format.
type Phone struct {
	Rel    string `json:"rel,omitempty"`
	Number string `json:"$t,omitempty"`
}
