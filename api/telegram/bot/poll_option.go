// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// PollOption contains information about one answer option in a poll.
type PollOption struct {
	// Option text, 1-100 characters.
	Text string `json:"text"`

	// Number of users that voted for this option.
	VoterCount int `json:"voter_count"`
}
