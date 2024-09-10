// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// PollAnswer represents an answer of a user in a non-anonymous poll.
type PollAnswer struct {
	// Unique poll identifier
	PollID string `json:"poll_id"`

	// The user, who changed the answer to the poll.
	User *User `json:"user"`

	// 0-based identifiers of answer options, chosen by the user. May be
	// empty if the user retracted their vote.
	OptionIDs []int `json:"option_ids"`
}
