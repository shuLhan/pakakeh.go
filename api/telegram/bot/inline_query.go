// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// InlineQuery represents an incoming inline query.
// When the user sends an empty query, your bot could return some default or
// trending results.
//
type InlineQuery struct {
	ID    string `json:"id"`    // Unique identifier for this qery
	From  *User  `json:"from"`  // Sender
	Query string `json:"query"` // Text of the query (up to 256 characters).

	// Offset of the results to be returned, can be controlled by the bot.
	Offset string `json:"offset"`

	// Optional. Sender location, only for bots that request user
	// location.
	Location *Location `json:"location"`
}
