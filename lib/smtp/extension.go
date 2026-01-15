// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package smtp

// Extension is an interface to implement extension for SMTP server.
type Extension interface {
	//
	// Name return the SMTP extension name to be used on reply of EHLO.
	//
	Name() string

	//
	// Params return the SMTP extension parameters.
	//
	Params() string

	//
	// ValidateCommand validate the command parameters, if the extension
	// provide custom parameters.
	//
	ValidateCommand(cmd *Command) error
}

var defaultExts = []Extension{
	&extDSN{},
}
