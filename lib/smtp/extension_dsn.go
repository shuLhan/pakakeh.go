// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package smtp

type extDSN struct {
}

// Name return the name of extension, which is "DSN".
func (dsn *extDSN) Name() string {
	return "DSN"
}

// Params return the SMTP extension parameters.
func (dsn *extDSN) Params() string {
	return ""
}

// ValidateCommand validate command parameter for MAIL and RCPT.
func (dsn *extDSN) ValidateCommand(cmd *Command) (err error) {
	if cmd == nil {
		return nil
	}

	switch cmd.Kind {
	case CommandMAIL:
	case CommandRCPT:
	case CommandZERO:
		return nil
	}

	return nil
}
