// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// EncryptedCredentials contains data required for decrypting and
// authenticating EncryptedPassportElement.
type EncryptedCredentials struct {
	// Base64-encoded encrypted JSON-serialized data with unique user's
	// payload, data hashes and secrets required for
	// EncryptedPassportElement decryption and authentication.
	Data string `json:"data"`

	// Base64-encoded data hash for data authentication
	Hash string `json:"hash"`

	// Base64-encoded secret, encrypted with the bot's public RSA key,
	// required for data decryption.
	Secret string `json:"secret"`
}
