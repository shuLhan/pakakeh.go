// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Name define Google contact name format.
type Name struct {
	Prefix GD `json:"gd$namePrefix,omitempty"`
	First  GD `json:"gd$givenName,omitempty"`
	Middle GD `json:"gd$additionalName,omitempty"`
	Last   GD `json:"gd$familyName,omitempty"`
	Suffix GD `json:"gd$nameSuffix,omitempty"`
	Full   GD `json:"gd$fullName,omitempty"`
}
