// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

// Address format.
type Address struct {
	Rel         string `json:"rel,omitempty"`
	Full        GD     `json:"gd$formattedAddress,omitempty"`
	POBox       GD     `json:"gd$pobox,omitempty"`
	Street      GD     `json:"gd$street,omitempty"`
	City        GD     `json:"gd$city,omitempty"`
	StateOrProv GD     `json:"gd$region,omitempty"`
	PostalCode  GD     `json:"gd$postcode,omitempty"`
	Country     GD     `json:"gd$country,omitempty"`
}
