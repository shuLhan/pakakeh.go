// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
