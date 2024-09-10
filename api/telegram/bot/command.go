// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

import (
	"fmt"
)

// Command represents a bot command.
type Command struct {
	// Function that will be called when Bot receive the command.
	// Handler can read command and its arguments through Message.Command
	// and Message.CommandArgs.
	Handler UpdateHandler `json:"-"`

	// Text of the command, 1-32 characters. Can contain only lowercase
	// English letters, digits and underscores.
	Command string `json:"command"`

	// Description of the command, 3-256 characters.
	Description string `json:"description"`
}

// validate will return an error if command is not valid.
func (cmd *Command) validate() error {
	if len(cmd.Command) == 0 || len(cmd.Command) > 32 {
		return errCommandLength(cmd.Command)
	}
	for x := 0; x < len(cmd.Command); x++ {
		b := cmd.Command[x]
		if b >= 'a' && b <= 'z' {
			continue
		}
		if b >= '0' && b <= '9' {
			continue
		}
		if b == '_' {
			continue
		}
		return errCommandValue(cmd.Command)
	}
	if len(cmd.Description) < 3 || len(cmd.Description) > 256 {
		return errDescLength(cmd.Command)
	}
	if cmd.Handler == nil {
		return errHandlerNil(cmd.Command)
	}
	return nil
}

func errCommandLength(cmd string) error {
	return fmt.Errorf("%q: the Command length must be between 1-32 characters", cmd)
}

func errCommandValue(cmd string) error {
	return fmt.Errorf("%q: command can contain only lowercase English letter, digits, and underscores", cmd)
}

func errDescLength(cmd string) error {
	return fmt.Errorf("%q: the Description length must be between 3-256 characters", cmd)
}

func errHandlerNil(cmd string) error {
	return fmt.Errorf("%q: the Command's Handler is not set", cmd)
}
