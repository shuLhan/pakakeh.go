// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// PollOption contains information about one answer option in a poll.
type PollOption struct {
	// Option text, 1-100 characters.
	Text string `json:"text"`

	// Number of users that voted for this option.
	VoterCount int `json:"voter_count"`
}
