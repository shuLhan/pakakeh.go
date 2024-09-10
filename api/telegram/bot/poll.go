// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Poll contains information about a poll.
type Poll struct {
	// Unique poll identifier
	ID string `json:"id"`

	// Poll type, currently can be “regular” or “quiz”.
	Type string `json:"type"`

	// Poll question, 1-255 characters.
	Question string `json:"question"`

	// List of poll options.
	Options []PollOption `json:"options"`

	// Optional. 0-based identifier of the correct answer option.
	// Available only for polls in the quiz mode, which are closed, or was
	// sent (not forwarded) by the bot or to the private chat with the
	// bot.
	CorrectOptionID int `json:"correct_option_id"`

	// Total number of users that voted in the poll.
	TotalVoterCount int `json:"total_voter_count"`

	// True, if the poll is closed.
	IsClosed bool `json:"is_closed"`

	// True, if the poll is anonymous.
	IsAnonymous bool `json:"is_anonymous"`

	// True, if the poll allows multiple answers.
	AllowsMultipleAnswers bool `json:"allow_multiple_answers"`
}
