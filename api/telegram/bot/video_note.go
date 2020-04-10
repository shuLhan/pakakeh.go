// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// VideoNote represents a video message (available in Telegram apps as of
// v.4.0).
//
type VideoNote struct {
	Document

	// Video width and height (diameter of the video message) as defined
	// by sender.
	Length int `json:"length"`

	// Duration of the video in seconds as defined by sender.
	Duration int `json:"duration"`
}
