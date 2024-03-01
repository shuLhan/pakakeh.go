// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/errors"
)

// List of errors.
var (
	ErrInvalidCredential = &errors.E{
		Code:    StatusInvalidCredential,
		Message: "5.7.8 Authentication credentials invalid",
	}

	errCmdSyntaxError = &errors.E{
		Code:    StatusCmdSyntaxError,
		Message: "Syntax error in parameter or arguments",
	}

	// See RFC 5321, section 4.5.3.1.9.  Treatment When Limits Exceeded
	errCmdTooLong = &errors.E{
		Code:    StatusCmdTooLong,
		Message: "Line too long",
	}
	errCmdUnknown = &errors.E{
		Code:    StatusCmdUnknown,
		Message: "Syntax error, command unknown",
	}
	// TODO:
	//	errInProcessing = &errors.E{
	//		Code:    StatusLocalError,
	//		Message: "Local error in processing",
	//	}

	errAuthMechanism = &errors.E{
		Code:    StatusParamUnimplemented,
		Message: "5.5.4 Command parameter not implemented",
	}
	errNotAuthenticated = &errors.E{
		Code:    StatusNotAuthenticated,
		Message: "5.7.0 Authentication required",
	}
	errBadSequence = &errors.E{
		Code:    StatusCmdBadSequence,
		Message: "Bad sequence of commands",
	}
)
