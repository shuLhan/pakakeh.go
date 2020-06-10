// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// InlineKeyboardButton represents one button of an inline keyboard. You must
// use exactly one of the optional fields.
//
type InlineKeyboardButton struct {
	// Label text on the button.
	Text string `json:"text"`

	// Optional. HTTP or tg:// url to be opened when button is pressed.
	URL string `json:"url"`

	// Optional. An HTTP URL used to automatically authorize the user. Can
	// be used as a replacement for the Telegram Login Widget.
	LoginURL *LoginURL `json:"login_url"`

	// Optional. Data to be sent in a callback query to the bot when
	// button is pressed, 1-64 bytes.
	CallbackData string `json:"callback_data"`

	// Optional. If set, pressing the button will prompt the user to
	// select one of their chats, open that chat and insert the bot‘s
	// username and the specified inline query in the input field. Can be
	// empty, in which case just the bot’s username will be inserted.
	SwitchInlineQuery string `json:"switch_inline_query"`

	// Optional. If set, pressing the button will insert the bot‘s
	// username and the specified inline query in the current chat's input
	// field. Can be empty, in which case only the bot’s username will be
	// inserted.
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat"`

	// Optional. Description of the game that will be launched when the
	// user presses the button.
	//
	// NOTE: This type of button must always be the first button in the
	// first row.
	CallbackGame *CallbackGame `json:"callback_game"`

	// Optional. Specify True, to send a Pay button.
	//
	// NOTE: This type of button must always be the first button in the
	// first row.
	Pay bool `json:"pay"`
}
