// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Video represents a video file.
type Video struct {
	Document

	Width  int `json:"width"`  // Video width as defined by sender.
	Height int `json:"height"` // Video height as defined by sender.

	// Duration of the video in seconds as defined by sender.
	Duration int `json:"duration"`
}
