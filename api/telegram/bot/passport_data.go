// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// PassportData contains information about Telegram Passport data shared with
// the bot by the user.
type PassportData struct {
	// Encrypted credentials required to decrypt the data.
	Credentials EncryptedCredentials
}
