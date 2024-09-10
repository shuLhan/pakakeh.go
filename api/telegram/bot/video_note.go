// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// VideoNote represents a video message (available in Telegram apps as of
// v.4.0).
type VideoNote struct {
	Document

	// Video width and height (diameter of the video message) as defined
	// by sender.
	Length int `json:"length"`

	// Duration of the video in seconds as defined by sender.
	Duration int `json:"duration"`
}
