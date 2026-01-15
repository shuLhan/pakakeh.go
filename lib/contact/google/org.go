// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Org as organisation.
type Org struct {
	Type     string `json:"rel,omitempty"`
	Name     GD     `json:"gd$orgName,omitempty"`
	JobTitle GD     `json:"gd$orgTitle,omitempty"`
}
