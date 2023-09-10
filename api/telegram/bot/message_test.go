// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
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
