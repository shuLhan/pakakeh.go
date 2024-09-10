// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Update an object represents an incoming update.
// At most one of the optional parameters can be present in any given update.
type Update struct {
	// Optional. New incoming message of any kind — text, photo, sticker,
	// etc.
	Message *Message `json:"Message"`

	// Optional. New version of a message that is known to the bot and was
	// edited.
	EditedMessage *Message `json:"edited_message"`

	// Optional. New incoming channel post of any kind — text, photo,
	// sticker, etc..
	ChannelPost *Message `json:"channel_post"`

	// Optional. New version of a channel post that is known to the bot
	// and was edited.
	EditedChannelPost *Message `json:"edited_channel_post"`

	// Optional. New incoming inline query.
	InlineQuery *InlineQuery `json:"inline_query"`

	// Optional. The result of an inline query that was chosen by a user
	// and sent to their chat partner.
	// Please see our documentation on the feedback collecting for details
	// on how to enable these updates for your bot.
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result"`

	// Optional. New incoming callback query.
	CallbackQuery *CallbackQuery `json:"callback_query"`

	// Optional. New incoming shipping query. Only for invoices with
	// flexible price.
	ShippingQuery *ShippingQuery `json:"shipping_query"`

	// Optional. New incoming pre-checkout query. Contains full
	// information about checkout.
	PreCheckoutQuery *PreCheckoutQuery `json:"pre_checkout_query"`

	// Optional. New poll state. Bots receive only updates about stopped
	// polls and polls, which are sent by the bot.
	Poll *Poll `json:"poll"`

	// Optional. A user changed their answer in a non-anonymous poll. Bots
	// receive new votes only in polls that were sent by the bot itself.
	PollAnswer *PollAnswer `json:"poll_answer"`

	// The update‘s unique identifier.
	// Update identifiers start from a certain positive number and
	// increase sequentially.
	// This ID becomes especially handy if you’re using Webhooks, since it
	// allows you to ignore repeated updates or to restore the correct
	// update sequence, should they get out of order.
	// If there are no new updates for at least a week, then identifier of
	// the next update will be chosen randomly instead of sequentially.
	ID int64 `json:"update_id"`
}
