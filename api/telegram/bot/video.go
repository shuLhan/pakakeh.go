// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

//
// Video represents a video file.
//
type Video struct {
	Document

	Width  int `json:"width"`  // Video width as defined by sender.
	Height int `json:"height"` // Video height as defined by sender.

	// Duration of the video in seconds as defined by sender.
	Duration int `json:"duration"`
}
