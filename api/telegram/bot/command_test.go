// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestCommand_validate(t *testing.T) {
	s33 := string(ascii.Random([]byte(ascii.Letters), 33))

	cases := []struct {
		cmd  Command
		exp  error
		desc string
	}{{
		desc: "with empty command",
		cmd:  Command{},
		exp:  errCommandLength(""),
	}, {
		desc: "with invalid command character '!'",
		cmd: Command{
			Command:     "a!",
			Description: "1234",
		},
		exp: errCommandValue("a!"),
	}, {
		desc: "with uppercase",
		cmd: Command{
			Command:     "Help",
			Description: string(ascii.Random([]byte(ascii.Letters), 257)),
		},
		exp: errCommandValue("Help"),
	}, {
		desc: "with command too long",
		cmd: Command{
			Command:     s33,
			Description: "1234",
		},
		exp: errCommandLength(s33),
	}, {
		desc: "with description too short",
		cmd: Command{
			Command:     "help",
			Description: "12",
		},
		exp: errDescLength("help"),
	}, {
		desc: "with description too long",
		cmd: Command{
			Command:     "help",
			Description: string(ascii.Random([]byte(ascii.Letters), 257)),
		},
		exp: errDescLength("help"),
	}, {
		desc: "Perfect",
		cmd: Command{
			Command:     "help",
			Description: "Bantuan",
			Handler:     func(_ Update) {},
		},
	}}

	for _, c := range cases {
		test.Assert(t, c.desc, c.exp, c.cmd.validate())
	}
}
