// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// ShippingQuery contains information about an incoming shipping query.
type ShippingQuery struct {
	// Unique query identifier.
	ID string `json:"id"`

	// User who sent the query.
	From *User `json:"from"`

	// Bot specified invoice payload.
	InvoicePayload string `json:"invoice_payload"`

	// User specified shipping address.
	ShippingAddress *ShippingAddress `json:"shipping_address"`
}
