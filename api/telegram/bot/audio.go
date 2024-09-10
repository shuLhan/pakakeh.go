// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Audio represents an audio file to be treated as music by the Telegram
// clients.
type Audio struct {
	// Optional. Performer of the audio as defined by sender or by audio
	// tags.
	Performer string `json:"performer"`

	// Optional. Title of the audio as defined by sender or by audio tags.
	Title string `json:"title"`

	Document

	// Duration of the audio in seconds as defined by sender.
	Duration int `json:"duration"`
}
