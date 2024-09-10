// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Sticker represents a sticker.
type Sticker struct {
	// Optional. For mask stickers, the position where the mask should be
	// placed.
	MaskPosition *MaskPosition `json:"mask_position"`

	// Optional. Emoji associated with the sticker.
	Emoji string `json:"emoji"`

	// Optional. Name of the sticker set to which the sticker belongs.
	SetName string `json:"set_name"`

	Document

	// Sticker width.
	Width int `json:"width"`

	// Sticker height.
	Height int `json:"height"`

	// True, if the sticker is animated.
	IsAnimated bool `json:"is_animated"`
}
