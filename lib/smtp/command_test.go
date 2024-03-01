// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestUnpack(t *testing.T) {
	cases := []struct {
		expCmd *Command
		expErr error
		desc   string
		b      string
	}{{
		desc:   "With invalid length",
		b:      "DAT\r\n",
		expErr: errCmdUnknown,
	}, {
		desc:   "With unknown command",
		b:      "DAT\r\n",
		expErr: errCmdUnknown,
	}, {
		desc:   "With length too long",
		b:      "VRFY " + string(ascii.Random([]byte(ascii.Letters), 513)),
		expErr: errCmdTooLong,
	}, {
		desc:   "Without CRLF",
		b:      "VRFY local.part@domain",
		expErr: errCmdSyntaxError,
	}, {
		desc: "DATA command",
		b:    "DATA\r\n",
		expCmd: &Command{
			Kind: CommandDATA,
		},
	}, {
		desc: "EHLO without argument",
		b:    "EHLO\r\n",
		expCmd: &Command{
			Kind: CommandEHLO,
		},
	}, {
		desc: "EHLO with empty argument",
		b:    "EHLO \r\n",
		expCmd: &Command{
			Kind: CommandEHLO,
		},
	}, {
		desc: "EHLO with argument",
		b:    "EHLO domain.com\r\n",
		expCmd: &Command{
			Kind: CommandEHLO,
			Arg:  "domain.com",
		},
	}, {
		desc:   "EXPN without argument",
		b:      "EXPN\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc:   "EXPN with empty argument",
		b:      "EXPN \r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc: "EXPN with argument",
		b:    "EXPN mailing-list\r\n",
		expCmd: &Command{
			Kind: CommandEXPN,
			Arg:  "mailing-list",
		},
	}, {
		desc: "HELO without argument",
		b:    "HELO\r\n",
		expCmd: &Command{
			Kind: CommandHELO,
		},
	}, {
		desc: "HELP with empty argument",
		b:    "HELO \r\n",
		expCmd: &Command{
			Kind: CommandHELO,
		},
	}, {
		desc: "HELO with argument",
		b:    "HELO domain.com\r\n",
		expCmd: &Command{
			Kind: CommandHELO,
			Arg:  "domain.com",
		},
	}, {
		desc: "HELP with empty argument",
		b:    "HELP\r\n",
		expCmd: &Command{
			Kind: CommandHELP,
		},
	}, {
		desc: "HELP with space",
		b:    "HELP \r\n",
		expCmd: &Command{
			Kind: CommandHELP,
		},
	}, {
		desc: "HELP with argument",
		b:    "HELP vrfy\r\n",
		expCmd: &Command{
			Kind: CommandHELP,
			Arg:  "vrfy",
		},
	}, {
		desc:   "MAIL with empty argument",
		b:      "MAIL FROM:\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc:   "MAIL with invalid command",
		b:      "MAIL FRO:<mail@box.com>\r\n",
		expErr: errCmdUnknown,
	}, {
		desc:   "MAIL with invalid path",
		b:      "MAIL FROM:<mail@box.com\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc:   "MAIL with invalid mailbox",
		b:      "MAIL FROM:<mail..@box.com>\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc:   "MAIL with space before path",
		b:      "MAIL FROM: <mail@box.com>\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc: "MAIL with domain only",
		b:    "MAIL FROM:<local@domain.com>\r\n",
		expCmd: &Command{
			Kind: CommandMAIL,
			Arg:  "local@domain.com",
		},
	}, {
		desc: "MAIL with source",
		b:    "MAIL FROM:<@domain:local@domain.com>\r\n",
		expCmd: &Command{
			Kind: CommandMAIL,
			Arg:  "local@domain.com",
		},
	}, {
		desc: "MAIL with param",
		b:    "MAIL FROM:<local@domain.com> key=value\r\n",
		expCmd: &Command{
			Kind: CommandMAIL,
			Arg:  "local@domain.com",
			Params: map[string]string{
				"key": "value",
			},
		},
	}, {
		desc: "MAIL with Params",
		b:    "MAIL FROM:<local@domain.com> key=value x=y\r\n",
		expCmd: &Command{
			Kind: CommandMAIL,
			Arg:  "local@domain.com",
			Params: map[string]string{
				"key": "value",
				"x":   "y",
			},
		},
	}, {
		desc: "MAIL with empty param value",
		b:    "MAIL FROM:<local@domain.com> key=value x=\r\n",
		expCmd: &Command{
			Kind: CommandMAIL,
			Arg:  "local@domain.com",
			Params: map[string]string{
				"key": "value",
			},
		},
	}, {
		desc: "NOOP",
		b:    "NOOP\r\n",
		expCmd: &Command{
			Kind: CommandNOOP,
		},
	}, {
		desc: "QUIT",
		b:    "QUIT\r\n",
		expCmd: &Command{
			Kind: CommandQUIT,
		},
	}, {
		desc:   "RCPT with empty argument",
		b:      "RCPT TO:\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc:   "RCPT with invalid command",
		b:      "RCPT:<mail@box.com>\r\n",
		expErr: errCmdUnknown,
	}, {
		desc:   "RCPT with space before path",
		b:      "RCPT TO: <mail@box.com>\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc: "RCPT with domain only",
		b:    "RCPT TO:<local@domain.com>\r\n",
		expCmd: &Command{
			Kind: CommandRCPT,
			Arg:  "local@domain.com",
		},
	}, {
		desc: "RCPT with postmaster",
		b:    "RCPT TO:<postmaster>\r\n",
		expCmd: &Command{
			Kind: CommandRCPT,
			Arg:  "postmaster",
		},
	}, {
		desc: "RCPT with source",
		b:    "RCPT TO:<@domain:local@domain.com>\r\n",
		expCmd: &Command{
			Kind: CommandRCPT,
			Arg:  "local@domain.com",
		},
	}, {
		desc: "RCPT with sources",
		b:    "RCPT TO:<@domain,@xyz:local@domain.com>\r\n",
		expCmd: &Command{
			Kind: CommandRCPT,
			Arg:  "local@domain.com",
		},
	}, {
		desc: "RCPT with param",
		b:    "RCPT TO:<local@domain.com> key=value\r\n",
		expCmd: &Command{
			Kind: CommandRCPT,
			Arg:  "local@domain.com",
			Params: map[string]string{
				"key": "value",
			},
		},
	}, {
		desc: "RCPT with Params",
		b:    "RCPT TO:<local@domain.com> key=value x=y\r\n",
		expCmd: &Command{
			Kind: CommandRCPT,
			Arg:  "local@domain.com",
			Params: map[string]string{
				"key": "value",
				"x":   "y",
			},
		},
	}, {
		desc: "RSET",
		b:    "RSET\r\n",
		expCmd: &Command{
			Kind: CommandRSET,
		},
	}, {
		desc:   "VRFY without argument",
		b:      "VRFY\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc:   "VRFY with empty argument",
		b:      "VRFY \r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc: "VRFY with argument",
		b:    "VRFY mail@box.com\r\n",
		expCmd: &Command{
			Kind: CommandVRFY,
			Arg:  "mail@box.com",
		},
	}, {
		desc:   "AUTH without argument",
		b:      "AUTH\r\n",
		expErr: errCmdSyntaxError,
	}, {
		desc: "AUTH with mechanism only",
		b:    "AUTH PLAIN\r\n",
		expCmd: &Command{
			Kind: CommandAUTH,
			Arg:  "PLAIN",
		},
	}, {
		desc: "AUTH with mechanism and initial-response",
		b:    "AUTH PLAIN AHRlc3QAMTIzNA==\r\n",
		expCmd: &Command{
			Kind:  CommandAUTH,
			Arg:   "PLAIN",
			Param: "AHRlc3QAMTIzNA==",
		},
	}}

	cmd := newCommand()
	for _, c := range cases {
		t.Log(c.desc)

		cmd.reset()
		err := cmd.unpack([]byte(c.b))
		if err != nil {
			test.Assert(t, "error", c.expErr, err)
			continue
		}

		test.Assert(t, "Command.Kind", c.expCmd.Kind, cmd.Kind)
		test.Assert(t, "Command.Arg", c.expCmd.Arg, cmd.Arg)
		test.Assert(t, "Command.Params", c.expCmd.Params, cmd.Params)
	}
}
