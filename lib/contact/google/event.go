// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Event format.
type Event struct {
	Rel  string    `json:"rel,omitempty"`
	When EventTime `json:"gd$when,omitempty"`
}
