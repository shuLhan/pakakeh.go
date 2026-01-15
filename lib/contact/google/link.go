// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Link define Google contact link type.
type Link struct {
	Rel  string `json:"rel,omitempty"`
	Type string `json:"type,omitempty"`
	HRef string `json:"href,omitempty"`
}
