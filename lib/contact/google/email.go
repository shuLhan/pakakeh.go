// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Email format.
type Email struct {
	Rel     string `json:"rel,omitempty"`
	Address string `json:"address,omitempty"`
	Primary string `json:"primary,omitempty"`
}
