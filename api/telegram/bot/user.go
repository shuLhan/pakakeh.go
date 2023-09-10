// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// User represents a Telegram user or bot.
type User struct {
	// User‘s or bot’s first name.
	FirstName string `json:"first_name"`

	// Optional. User‘s or bot’s last name.
	LastName string `json:"last_name"`

	// Optional. User‘s or bot’s username.
	Username string `json:"username"`

	// Optional. IETF language tag of the user's language.
	LanguageCode string `json:"language_code"`

	// Unique identifier for this user or bot
	ID int `json:"id"`

	// True, if this user is a bot
	IsBot bool `json:"is_bot"`

	// Optional. True, if the bot can be invited to groups. Returned only
	// in getMe.
	CanJoinGroups bool `json:"can_join_groups"`

	// Optional. True, if privacy mode is disabled for the bot. Returned
	// only in getMe.
	CanReadAllGroupMessages bool `json:"can_read_all_group_messages"`

	// Optional. True, if the bot supports inline queries. Returned only
	// in getMe.
	SupportsInlineQueries bool `json:"supports_inline_queries"`
}
