// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Root define the root of Google's contact in JSON.
type Root struct {
	Version  string `json:"version,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	Feed     Feed   `json:"feed,omitempty"`
}
