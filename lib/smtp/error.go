// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"github.com/shuLhan/share/lib/errors"
)

// List of errors.
var (
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
	errInProcessing = &errors.E{
		Code:    StatusLocalError,
		Message: "Local error in processing",
	}
)
