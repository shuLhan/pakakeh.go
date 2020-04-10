// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// List of chat types.
const (
	ChatTypeChannel    = "channel"
	ChatTypeGroup      = "group"
	ChatTypePrivate    = "private"
	ChatTypeSupergroup = "supergroup"
)

//
// Chat represents a chat.
//
type Chat struct {

	// Unique identifier for this chat.
	//
	// This number may be greater than 32 bits and some programming
	// languages may have difficulty/silent defects in interpreting it.
	// But it is smaller than 52 bits, so a signed 64 bit integer or
	// double-precision float type are safe for storing this identifier.
	ID int64 `json:"id"`

	// Type of chat, can be either “private”, “group”, “supergroup” or
	// “channel”.
	Type string `json:"type"`

	// Optional. Title, for supergroups, channels and group chats.
	Title string `json:"title"`

	// Optional. Username, for private chats, supergroups and channels if
	// available.
	Username string `json:"username"`

	// Optional. First name of the other party in a private chat.
	FirstName string `json:"first_name"`

	// Optional. Last name of the other party in a private chat.
	LastName string `json:"last_name"`

	// Optional. Chat photo. Returned only in getChat.
	Photo *ChatPhoto `json:"chat_photo"`

	// Optional. Description, for groups, supergroups and channel chats.
	// Returned only in getChat.
	Description string `json:"description"`

	// Optional. Chat invite link, for groups, supergroups and channel
	// chats.
	// Each administrator in a chat generates their own invite links, so
	// the bot must first generate the link using exportChatInviteLink.
	// Returned only in getChat.
	InviteLink string `json:"invite_link"`

	// Optional. Pinned message, for groups, supergroups and channels.
	// Returned only in getChat.
	PinnedMesage *Message `json:"pinned_mesage"`

	// Optional. Default chat member permissions, for groups and
	// supergroups. Returned only in getChat.
	Permissions *ChatPermissions `json:"permissions"`

	// Optional. For supergroups, the minimum allowed delay between
	// consecutive messages sent by each unpriviledged user. Returned only
	// in getChat.
	SlowModeDelay int `json:"slow_mode_delay"`

	// Optional. For supergroups, name of group sticker set. Returned only
	// in getChat.
	StickerSetName string `json:"sticker_set_name"`

	// Optional. True, if the bot can change the group sticker set.
	// Returned only in getChat.
	CanSetStickerSet bool `json:"can_set_sticker_set"`
}
