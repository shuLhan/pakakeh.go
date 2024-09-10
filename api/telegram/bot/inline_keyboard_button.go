// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// InlineKeyboardButton represents one button of an inline keyboard. You must
// use exactly one of the optional fields.
type InlineKeyboardButton struct {
	// Optional. An HTTP URL used to automatically authorize the user. Can
	// be used as a replacement for the Telegram Login Widget.
	LoginURL *LoginURL `json:"login_url"`

	// Optional. Description of the game that will be launched when the
	// user presses the button.
	//
	// NOTE: This type of button must always be the first button in the
	// first row.
	CallbackGame *CallbackGame `json:"callback_game"`

	// Label text on the button.
	Text string `json:"text"`

	// Optional. HTTP or tg:// url to be opened when button is pressed.
	URL string `json:"url"`

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

	// Optional. Specify True, to send a Pay button.
	//
	// NOTE: This type of button must always be the first button in the
	// first row.
	Pay bool `json:"pay"`
}
