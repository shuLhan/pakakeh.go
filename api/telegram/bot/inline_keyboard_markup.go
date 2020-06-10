// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// InlineKeyboardMarkup represents an inline keyboard that appears right next
// to the message it belongs to.
//
type InlineKeyboardMarkup struct {
	// Array of button rows, each represented by an Array of
	// InlineKeyboardButton objects.
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}
