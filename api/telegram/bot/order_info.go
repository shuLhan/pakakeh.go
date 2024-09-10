// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// OrderInfo represents information about an order.
type OrderInfo struct {
	// Optional. User shipping address
	ShippingAddress *ShippingAddress `json:"shipping_address"`

	// Optional. User name
	Name string `json:"name"`

	// Optional. User's phone number
	PhoneNumber string `json:"phone_number"`

	// Optional. User email
	Email string `json:"email"`
}
