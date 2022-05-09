// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// Sticker represents a sticker.
type Sticker struct {
	Document

	// Sticker width.
	Width int `json:"width"`

	// Sticker height.
	Height int `json:"height"`

	// Optional. Emoji associated with the sticker.
	Emoji string `json:"emoji"`

	// Optional. Name of the sticker set to which the sticker belongs.
	SetName string `json:"set_name"`

	// Optional. For mask stickers, the position where the mask should be
	// placed.
	MaskPosition *MaskPosition `json:"mask_position"`

	// True, if the sticker is animated.
	IsAnimated bool `json:"is_animated"`
}
