// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// EncryptedCredentials
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
