// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package exec wrap the standar package "os/exec" to simplify calling Run
// with stdout and stderr.
//
package exec

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

//
// ParseCommandArgs parse the input string into command and arguments.
// This function detect possible single, double, or back quote on arguments.
//
func ParseCommandArgs(in string) (cmd string, args []string) {
	var (
		quote   rune
		cmdArgs []string
	)

	sb := new(strings.Builder)

	for _, r := range in {
		if quote > 0 {
			if r == quote {
				arg := sb.String()
				if len(arg) > 0 {
					cmdArgs = append(cmdArgs, sb.String())
				}

				sb.Reset()
				quote = 0
			} else {
				sb.WriteRune(r)
			}
			continue
		}
		if r == '\'' || r == '"' || r == '`' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' {
			arg := sb.String()
			if len(arg) > 0 {
				cmdArgs = append(cmdArgs, sb.String())
			}
			sb.Reset()
			continue
		}
		sb.WriteRune(r)
	}

	arg := sb.String()
	if len(arg) > 0 {
		cmdArgs = append(cmdArgs, sb.String())
	}
	sb.Reset()

	if len(cmdArgs) > 0 {
		cmd = cmdArgs[0]
	}
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}

	return cmd, args
}

//
// Run the command and arguments in the string cmd.
// If stdout or stderr is nil, it will default to os.Stdout and/or os.Stderr.
//
func Run(command string, stdout, stderr io.Writer) (err error) {
	cmd, args := ParseCommandArgs(command)

	execCmd := exec.Command(cmd, args...)

	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	return execCmd.Run()
}
