// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// MessageForward define the content for forwarded message.
type MessageForward struct {
	// Optional. For forwarded messages, sender of the original message.
	ForwardFrom *User `json:"forward_from"`

	// Optional. For messages forwarded from channels, information about
	// the original channel.
	ForwardChat *Chat `json:"forward_from_chat"`

	// Optional. For messages forwarded from channels, signature of the post
	// author if present.
	ForwardSignature string `json:"forward_signature"`

	// Optional. Sender's name for messages forwarded from users who
	// disallow adding a link to their account in forwarded messages.
	ForwardSenderName string `json:"forward_sender_name"`

	// Optional. For messages forwarded from channels, identifier of the
	// original message in the channel.
	ForwardID int64 `json:"forward_from_message_id"`

	// Optional. For forwarded messages, date the original message was
	// sent in Unix time.
	ForwardDate int64 `json:"forward_date"`
}
