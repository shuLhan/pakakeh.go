// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Generator define Google contact generator.
type Generator struct {
	Version string `json:"version,omitempty"`
	URI     string `json:"uri,omitempty"`
	Value   string `json:"$t,omitempty"`
}
