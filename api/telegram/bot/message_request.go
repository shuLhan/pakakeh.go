// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// messageRequest represents internal message to be used on sendMessage
//
type messageRequest struct {
	// Unique identifier for the target chat or username of the target
	// channel (in the format @channelusername).
	ChatID interface{} `json:"chat_id"`

	// Text of the message to be sent, 1-4096 characters after entities
	// parsing.
	Text string `json:"text"`

	// Send Markdown or HTML, if you want Telegram apps to show bold,
	// italic, fixed-width text or inline URLs in your bot's message.
	ParseMode string `json:"parse_mode,omitempty"`

	// Disables link previews for links in this message.
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`

	// Sends the message silently. Users will receive a notification with
	// no sound.
	DisableNotification bool `json:"disable_notification,omitempty"`

	// If the message is a reply, ID of the original message.
	ReplyToMessageID int64 `json:"reply_to_message_id,omitempty"`

	// Additional interface options. A JSON-serialized object for an
	// inline keyboard, custom reply keyboard, instructions to remove
	// reply keyboard or to force a reply from the user.
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}
