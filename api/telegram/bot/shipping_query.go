// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// ShippingQuery contains information about an incoming shipping query.
type ShippingQuery struct {
	// User who sent the query.
	From *User `json:"from"`

	// User specified shipping address.
	ShippingAddress *ShippingAddress `json:"shipping_address"`

	// Unique query identifier.
	ID string `json:"id"`

	// Bot specified invoice payload.
	InvoicePayload string `json:"invoice_payload"`
}
