// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// EncryptedPassportElement contains information about documents or other
// Telegram Passport elements shared with the bot by the user.
//
type EncryptedPassportElement struct {
	//
	// Element type. One of “personal_details”, “passport”,
	// “driver_license”, “identity_card”, “internal_passport”, “address”,
	// “utility_bill”, “bank_statement”, “rental_agreement”,
	// “passport_registration”, “temporary_registration”, “phone_number”,
	// “email”.
	//
	Type string `json:"type"`

	//
	// Optional. Base64-encoded encrypted Telegram Passport element data
	// provided by the user, available for “personal_details”, “passport”,
	// “driver_license”, “identity_card”, “internal_passport” and
	// “address” types.
	// Can be decrypted and verified using the accompanying
	// EncryptedCredentials.
	//
	Data string `json:"data"`

	//
	// Optional. User's verified phone number, available only for
	// “phone_number” type.
	//
	PhoneNumber string `json:"phone_number"`

	//
	// Optional. User's verified email address, available only for “email”
	// type.
	//
	Email string `json:"email"`

	//
	// Optional. Array of encrypted files with documents provided by the
	// user, available for “utility_bill”, “bank_statement”,
	// “rental_agreement”, “passport_registration” and
	// “temporary_registration” types.
	// Files can be decrypted and verified using the accompanying
	// EncryptedCredentials.
	//
	Files []PassportFile `json:"files"`

	//
	// Optional. Encrypted file with the front side of the document,
	// provided by the user.
	// Available for “passport”, “driver_license”, “identity_card” and
	// “internal_passport”. The file can be decrypted and verified using
	// the accompanying EncryptedCredentials.
	//
	FrontSide *PassportFile `json:"front_size"`

	//
	// Optional. Encrypted file with the reverse side of the document,
	// provided by the user.
	// Available for “driver_license” and “identity_card”.
	// The file can be decrypted and verified using the accompanying
	// EncryptedCredentials.
	//
	ReverseSide *PassportFile `json:"reverse_side"`

	//
	// Optional. Encrypted file with the selfie of the user holding a
	// document, provided by the user; available for “passport”,
	// “driver_license”, “identity_card” and “internal_passport”.
	// The file can be decrypted and verified using the accompanying
	// EncryptedCredentials.
	//
	Selfie *PassportFile `json:"selfie"`

	//
	// Optional. Array of encrypted files with translated versions of
	// documents provided by the user.
	// Available if requested for “passport”, “driver_license”,
	// “identity_card”, “internal_passport”, “utility_bill”,
	// “bank_statement”, “rental_agreement”, “passport_registration” and
	// “temporary_registration” types.
	//
	// Files can be decrypted and verified using the accompanying
	// EncryptedCredentials.
	//
	Translation []PassportFile `json:"translation"`

	// Base64-encoded element hash for using in
	// PassportElementErrorUnspecified
	Hash string `json:"hash"`
}
