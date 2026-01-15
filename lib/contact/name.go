// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package contact

// Name define contact's name.
type Name struct {
	Given       string `json:"givenName"`
	Middle      string `json:"middleName"`
	Family      string `json:"familyName"`
	Prefix      string `json:"prefix"`
	Suffix      string `json:"suffix"`
	GivenSound  string `json:"givenNameSound"`
	FamilySound string `json:"familyNameSound"`
}
