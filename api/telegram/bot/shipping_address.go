// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// ShippingAddress represents a shipping address.
//
type ShippingAddress struct {
	// ISO 3166-1 alpha-2 country code.
	CountryCode string `json:"country_code"`

	// State, if applicable.
	State string `json:"state"`

	// City.
	City string `json:"city"`

	// First line for the address.
	StreetLine1 string `json:"street_line1"`

	// Second line for the address.
	StreetLine2 string `json:"street_line2"`

	// Address post code.
	PostCode string `json:"post_code"`
}
