// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// InlineKeyboardMarkup represents an inline keyboard that appears right next
// to the message it belongs to.
type InlineKeyboardMarkup struct {
	// Array of button rows, each represented by an Array of
	// InlineKeyboardButton objects.
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}
