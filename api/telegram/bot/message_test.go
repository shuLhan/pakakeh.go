// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestMessage_parseCommandArgs(t *testing.T) {
	cases := []struct {
		expCommand string
		expArgs    string
		msg        Message
	}{{
		msg: Message{
			Text: "Definisi /analisis",
			Entities: []MessageEntity{{
				Type:   EntityTypeBotCommand,
				Offset: 9,
				Length: 9,
			}},
		},
		expCommand: "analisis",
	}, {
		msg: Message{
			Text: "/definisi analisis",
			Entities: []MessageEntity{{
				Type:   EntityTypeBotCommand,
				Offset: 0,
				Length: 9,
			}},
		},
		expCommand: "definisi",
		expArgs:    "analisis",
	}, {
		msg: Message{
			Text: "/definisi@KamuskuBot analisis",
			Entities: []MessageEntity{{
				Type:   EntityTypeBotCommand,
				Offset: 0,
				Length: 20,
			}},
		},
		expCommand: "definisi",
		expArgs:    "analisis",
	}}

	for _, c := range cases {
		c.msg.parseCommandArgs()

		test.Assert(t, "Command", c.expCommand, c.msg.Command)
		test.Assert(t, "CommandArgs", c.expArgs, c.msg.CommandArgs)
	}
}
