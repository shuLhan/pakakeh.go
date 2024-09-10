// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// InlineQuery represents an incoming inline query.
// When the user sends an empty query, your bot could return some default or
// trending results.
type InlineQuery struct {
	From *User `json:"from"` // Sender

	// Optional. Sender location, only for bots that request user
	// location.
	Location *Location `json:"location"`

	ID    string `json:"id"`    // Unique identifier for this qery
	Query string `json:"query"` // Text of the query (up to 256 characters).

	// Offset of the results to be returned, can be controlled by the bot.
	Offset string `json:"offset"`
}
