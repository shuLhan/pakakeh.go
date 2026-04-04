// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Address format.
type Address struct {
	Rel         string `json:"rel,omitempty"`
	Full        GD     `json:"gd$formattedAddress"`
	POBox       GD     `json:"gd$pobox"`
	Street      GD     `json:"gd$street"`
	City        GD     `json:"gd$city"`
	StateOrProv GD     `json:"gd$region"`
	PostalCode  GD     `json:"gd$postcode"`
	Country     GD     `json:"gd$country"`
}
