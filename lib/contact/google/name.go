// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Name define Google contact name format.
type Name struct {
	Prefix GD `json:"gd$namePrefix"`
	First  GD `json:"gd$givenName"`
	Middle GD `json:"gd$additionalName"`
	Last   GD `json:"gd$familyName"`
	Suffix GD `json:"gd$nameSuffix"`
	Full   GD `json:"gd$fullName"`
}
