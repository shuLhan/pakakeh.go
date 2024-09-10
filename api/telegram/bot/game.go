// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Game represents a game.
// Use BotFather to create and edit games, their short names will act as
// unique identifiers.
type Game struct {
	// Optional. Animation that will be displayed in the game message in
	// chats. Upload via BotFather.
	Animation *Animation `json:"animation"`

	Title       string `json:"title"`       // Title of the game.
	Description string `json:"description"` // Description of the game.

	// Optional. Brief description of the game or high scores included in
	// the game message.
	// Can be automatically edited to include current high scores for the
	// game when the bot calls setGameScore, or manually edited using
	// editMessageText.
	// 0-4096 characters.
	Text string `json:"text"`

	// Photo that will be displayed in the game message in chats.
	Photo []PhotoSize `json:"photo"`

	// Optional. Special entities that appear in text, such as usernames,
	// URLs, bot commands, etc.
	TextEntities []MessageEntity `json:"text_entities"`
}
