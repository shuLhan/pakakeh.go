// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

package pasetov2

// JSONFooter define the optional metadata and data at the footer of the
// token that are not included in signature.
type JSONFooter struct {
	Data map[string]any `json:"data,omitempty"`
	KID  string         `json:"kid"`
}
