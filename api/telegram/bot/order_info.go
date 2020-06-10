// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// OrderInfo represents information about an order.
//
type OrderInfo struct {
	// Optional. User name
	Name string `json:"name"`

	// Optional. User's phone number
	PhoneNumber string `json:"phone_number"`

	// Optional. User email
	Email string `json:"email"`

	// Optional. User shipping address
	ShippingAddress *ShippingAddress `json:"shipping_address"`
}
