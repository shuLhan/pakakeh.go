// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// PreCheckoutQuery contains information about an incoming pre-checkout query.
//
type PreCheckoutQuery struct {
	// Unique query identifier.
	ID string `json:"id"`

	// User who sent the query.
	From *User `json:"from"`

	// Three-letter ISO 4217 currency code.
	Currency string `json:"currency"`

	// Total price in the smallest units of the currency (integer, not
	// float/double).
	// For example, for a price of US$ 1.45 pass amount = 145.
	// See the exp parameter in currencies.json, it shows the number
	// of digits past the decimal point for each currency (2 for the
	// majority of currencies).
	TotalAmount int `json:"total_amount"`

	// Bot specified invoice payload.
	InvoicePayload string `json:"invoice_payload"`

	// Optional. Identifier of the shipping option chosen by the user.
	ShippingOptionID string `json:"shipping_option_id"`

	// Optional. Order info provided by the user.
	OrderInfo *OrderInfo `json:"order_info"`
}
