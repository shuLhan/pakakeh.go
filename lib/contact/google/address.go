// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

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
