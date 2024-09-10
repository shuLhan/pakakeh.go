// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Contact represents a phone contact.
type Contact struct {
	PhoneNumber string `json:"phone_number"` // Contact's phone number.
	FirstName   string `json:"first_name"`   // Contact's first name.

	// Optional. Contact's last name.
	LastName string `json:"last_name"`

	// Optional. Additional data about the contact in the form of a vCard
	VCard string `json:"vcard"`

	// Optional. Contact's user identifier in Telegram
	UserID int64 `json:"user_id"`
}
