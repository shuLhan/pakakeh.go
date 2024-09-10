// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// PreCheckoutQuery contains information about an incoming pre-checkout query.
type PreCheckoutQuery struct {
	// User who sent the query.
	From *User `json:"from"`

	// Optional. Order info provided by the user.
	OrderInfo *OrderInfo `json:"order_info"`

	// Unique query identifier.
	ID string `json:"id"`

	// Three-letter ISO 4217 currency code.
	Currency string `json:"currency"`

	// Bot specified invoice payload.
	InvoicePayload string `json:"invoice_payload"`

	// Optional. Identifier of the shipping option chosen by the user.
	ShippingOptionID string `json:"shipping_option_id"`

	// Total price in the smallest units of the currency (integer, not
	// float/double).
	// For example, for a price of US$ 1.45 pass amount = 145.
	// See the exp parameter in currencies.json, it shows the number
	// of digits past the decimal point for each currency (2 for the
	// majority of currencies).
	TotalAmount int `json:"total_amount"`
}
