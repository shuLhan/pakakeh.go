// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

// Package exec wrap the standar package "os/exec" to simplify calling Run
// with stdout and stderr.
package exec

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// ParseCommandArgs parse the input string into command and arguments.
// This function detect single, double, or back quote on arguments; and
// escaped spaces using backslash.
func ParseCommandArgs(in string) (cmd string, args []string) {
	var (
		prev    rune
		quote   rune
		cmdArgs []string
	)

	sb := new(strings.Builder)

	for _, r := range in {
		if quote > 0 {
			switch r {
			case quote:
				if prev == '\\' {
					sb.WriteRune(r)
					prev = r
				} else {
					arg := sb.String()
					if len(arg) > 0 {
						cmdArgs = append(cmdArgs, sb.String())
					}
					sb.Reset()
					quote = 0
				}
			case '\\':
				if prev == '\\' {
					sb.WriteRune(r)
					prev = 0
				} else {
					prev = r
				}
			default:
				if prev == '\\' {
					sb.WriteRune('\\')
				}
				sb.WriteRune(r)
				prev = r
			}
			continue
		}
		if r == '\'' || r == '"' || r == '`' {
			quote = r
			prev = r
			continue
		}
		if r == '\\' {
			if prev == '\\' {
				sb.WriteRune(r)
				prev = 0
			} else {
				prev = r
			}
			continue
		}
		if r == ' ' || r == '\t' {
			if prev == '\\' {
				sb.WriteRune(r)
			} else {
				arg := sb.String()
				if len(arg) > 0 {
					cmdArgs = append(cmdArgs, sb.String())
				}
				sb.Reset()
			}
			prev = r
			continue
		}
		sb.WriteRune(r)
		prev = r
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

// Run the command and arguments in the string cmd.
// If stdout or stderr is nil, it will default to os.Stdout and/or os.Stderr.
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
