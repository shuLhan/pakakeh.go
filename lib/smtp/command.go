// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
)

type CommandKind int

// List of SMTP commands.
const (
	CommandZERO CommandKind = 0
	CommandHELO             = 1 << iota
	CommandEHLO
	CommandMAIL
	CommandRCPT
	CommandDATA
	CommandRSET
	CommandVRFY
	CommandEXPN
	CommandHELP
	CommandNOOP
	CommandQUIT
)

type Command struct {
	Kind   CommandKind
	Arg    string
	Params map[string]string
}

//
// newCommand create or get new command from pool.
//
func newCommand() *Command {
	return &Command{
		Params: make(map[string]string),
	}
}

//
// parsePath parse the reverse-path in MAIL command or forward-path in RCPT
// command, and their optional parameters.
//
// Syntax,
//
//	MAIL FROM:<[@domain[,...]:]local@domain> [ SP params ]
//
func (cmd *Command) parsePath(b []byte) error {
	if len(b) == 0 {
		return errCmdSyntaxError
	}

	if b[0] != '<' {
		return errCmdSyntaxError
	}

	x := len(b) - 1
	for ; x > 0; x-- {
		if b[x] == '>' {
			break
		}
	}
	if x == 0 {
		return errCmdSyntaxError
	}
	mb, err := ParsePath(b[:x+1])
	if err != nil {
		return errCmdSyntaxError
	}

	cmd.Arg = string(mb)
	if x < len(b) {
		err = cmd.parseParams(b[x+1:])
	}

	return err
}

//
// parseParams parse parameters in MAIL or RCPT argument.  The parameters have
// the following syntax,
//
//	key=value [ SP key=value ]
//
func (cmd *Command) parseParams(line []byte) error {
	var x int
	var k, v []byte

	for ; x < len(line); x++ {
		for ; x < len(line); x++ {
			if line[x] != ' ' {
				break
			}
		}
		for ; x < len(line); x++ {
			if line[x] == '=' {
				x++
				break
			}
			k = append(k, line[x])
		}
		if x == len(line) {
			break
		}
		for ; x < len(line); x++ {
			if line[x] == ' ' {
				break
			}
			v = append(v, line[x])
		}

		if len(k) > 0 && len(v) > 0 {
			if cmd.Params == nil {
				cmd.Params = make(map[string]string)
			}
			cmd.Params[string(k)] = string(v)
		}
		k = nil
		v = nil
	}

	return nil
}

//
// reset command fields to its zero value for re-use.
//
func (cmd *Command) reset() {
	cmd.Arg = ""
	cmd.Params = nil
}

//
// unpack parse a command type, argument, and their parameters.
//
func (cmd *Command) unpack(b []byte) (err error) { // nolint: gocyclo
	// Minimum command length is 4 + CRLF.
	if len(b) < 6 {
		return errCmdUnknown
	}
	if len(b) > 512 {
		return errCmdTooLong
	}

	// Remove CRLF.
	b = b[:len(b)-2]

	// Remove trailing spaces.
	b = bytes.TrimRight(b, " ")
	b = bytes.ToLower(b)

	switch b[0] {
	case 'd':
		if bytes.Equal([]byte("data"), b[0:4]) {
			cmd.Kind = CommandDATA
			return nil
		}

	case 'e':
		if bytes.Equal([]byte("ehlo"), b[0:4]) {
			cmd.Kind = CommandEHLO
			if len(b) > 5 {
				cmd.Arg = string(bytes.TrimSpace(b[5:]))
			}
			return nil
		}

		if bytes.Equal([]byte("expn"), b[0:4]) {
			cmd.Kind = CommandEXPN
			if len(b) > 5 {
				cmd.Arg = string(bytes.TrimSpace(b[5:]))
			}
			if len(cmd.Arg) == 0 {
				return errCmdSyntaxError
			}
			return nil
		}

	case 'h':
		if bytes.Equal([]byte("helo"), b[0:4]) {
			cmd.Kind = CommandHELO
			if len(b) > 5 {
				cmd.Arg = string(bytes.TrimSpace(b[5:]))
			}
			return nil
		}

		if bytes.Equal([]byte("help"), b[0:4]) {
			cmd.Kind = CommandHELP
			if len(b) > 5 {
				cmd.Arg = string(bytes.TrimSpace(b[5:]))
			}
			return nil
		}

	case 'm':
		if len(b) >= 10 && bytes.Equal([]byte("mail from:"), b[0:10]) {
			err = cmd.parsePath(b[10:])
			if err != nil {
				return err
			}
			cmd.Kind = CommandMAIL
			return nil
		}

	case 'n':
		if bytes.Equal([]byte("noop"), b[0:4]) {
			cmd.Kind = CommandNOOP
			return nil
		}

	case 'q':
		if bytes.Equal([]byte("quit"), b[0:4]) {
			cmd.Kind = CommandQUIT
			return nil
		}

	case 'r':
		if len(b) >= 8 && bytes.Equal([]byte("rcpt to:"), b[0:8]) {
			err = cmd.parsePath(b[8:])
			if err != nil {
				return err
			}
			cmd.Kind = CommandRCPT
			return nil
		}

		if bytes.Equal([]byte("rset"), b[0:4]) {
			cmd.Kind = CommandRSET
			return nil
		}

	case 'v':
		if bytes.Equal([]byte("vrfy"), b[0:4]) {
			cmd.Kind = CommandVRFY
			if len(b) > 5 {
				cmd.Arg = string(bytes.TrimSpace(b[5:]))
			}
			if len(cmd.Arg) == 0 {
				return errCmdSyntaxError
			}
			return nil
		}
	}
	return errCmdUnknown
}
