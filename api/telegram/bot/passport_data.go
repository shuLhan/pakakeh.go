// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// PassportData contains information about Telegram Passport data shared with
// the bot by the user.
type PassportData struct {
	// Encrypted credentials required to decrypt the data.
	Credentials EncryptedCredentials
}
