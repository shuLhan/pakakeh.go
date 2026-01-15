// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package contact

// Address define contact's address.
type Address struct {
	Type        string `json:"type,omitempty"`
	POBox       string `json:"pobox,omitempty"`
	Ext         string `json:"extension,omitempty"`
	Street      string `json:"street,omitempty"`
	City        string `json:"city,omitempty"`
	StateOrProv string `json:"stateOrProvince,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	Country     string `json:"country,omitempty"`
}
