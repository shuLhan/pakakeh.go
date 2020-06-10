// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// Audio represents an audio file to be treated as music by the Telegram
// clients.
type Audio struct {
	Document

	// Duration of the audio in seconds as defined by sender.
	Duration int `json:"duration"`

	// Optional. Performer of the audio as defined by sender or by audio
	// tags.
	Performer string `json:"performer"`

	// Optional. Title of the audio as defined by sender or by audio tags.
	Title string `json:"title"`
}
