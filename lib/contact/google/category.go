// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Category format.
type Category struct {
	Scheme string `json:"scheme,omitempty"`
	Term   string `json:"term,omitempty"`
}
