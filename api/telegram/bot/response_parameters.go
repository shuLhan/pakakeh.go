// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// responseParameters contains information about why a request was
// unsuccessful.
//
type responseParameters struct {
	// Optional. The group has been migrated to a supergroup with the
	// specified identifier. This number may be greater than 32 bits and
	// some programming languages may have difficulty/silent defects in
	// interpreting it. But it is smaller than 52 bits, so a signed 64 bit
	// integer or double-precision float type are safe for storing this
	// identifier.
	MigrateToChatID int `json:"migrate_to_chat_id"`

	// Optional. In case of exceeding flood control, the number of seconds
	// left to wait before the request can be repeated.
	RetryAfter int `json:"retry_after"`
}
