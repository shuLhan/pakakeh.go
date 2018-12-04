// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Extension is an interface to implement extension for SMTP server.
//
type Extension interface {
	//
	// Name return the SMTP extension name to be used on reply of EHLO.
	//
	Name() string

	//
	// ValidateCommand validate the command parameters, if the extension
	// provide custom parameters.
	//
	ValidateCommand(cmd *Command) error
}

var defaultExts = []Extension{
	&extDSN{},
}
