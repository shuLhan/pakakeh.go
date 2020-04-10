// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// Invoice contains basic information about an invoice.
//
type Invoice struct {
	Title       string `json:"title"`       // Product name
	Description string `json:"description"` // Product description

	// Unique bot deep-linking parameter that can be used to generate
	// this invoice.
	StartParameter string `json:"start_parameter"`

	// Three-letter ISO 4217 currency code
	Currency string `json:"currency"`

	// Total price in the smallest units of the currency (integer, not
	// float/double). For example, for a price of US$ 1.45 pass amount =
	// 145. See the exp parameter in currencies.json, it shows the number
	// of digits past the decimal point for each currency (2 for the
	// majority of currencies).
	TotalAmount int `json:"total_amount"`
}
