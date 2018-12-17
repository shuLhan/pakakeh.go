// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors provide an error type with code.
package errors

import (
	"net/http"
)

//
// E define custom error type with code.
//
type E struct {
	Code    int
	Message string
}

//
// Internal define an error for internal server.
//
func Internal() *E {
	return &E{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}
}

//
// InvalidInput generate an error for invalid input.
//
func InvalidInput(field string) *E {
	return &E{
		Code:    http.StatusBadRequest,
		Message: "Invalid input: " + field,
	}
}

//
// Error implement the error interface.
//
func (e *E) Error() string {
	return e.Message
}
