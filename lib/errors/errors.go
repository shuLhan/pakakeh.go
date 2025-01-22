// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package errors provide an error type with code.
package errors

import (
	"net/http"
	"reflect"
)

// E define custom error that wrap underlying error with custom code, message,
// and name.
//
// The Code field is required, used to communicate the HTTP response code.
// The Message field is optional, it's used to communicate the actual error
// message from server, to be readable by human.
// The Name field is optional, intended to be consumed by program, for
// example, to provide a key as translation of Message into user's locale
// defined language.
type E struct {
	err     error
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// Internal define an error caused by server.
func Internal(err error) *E {
	return &E{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Name:    "ERR_INTERNAL",
		err:     err,
	}
}

// InvalidInput generate an error for invalid input.
func InvalidInput(field string) *E {
	return &E{
		Code:    http.StatusBadRequest,
		Message: "invalid input: " + field,
		Name:    "ERR_INVALID_INPUT",
	}
}

// Error implement the error interface.
func (e *E) Error() string {
	return e.Message
}

// As set the target to e only if only target is **E.
func (e *E) As(target any) bool {
	_, ok := target.(**E)
	if ok {
		val := reflect.ValueOf(target)
		val.Elem().Set(reflect.ValueOf(e))
		return ok
	}
	return false
}

// Is return true if the target error is instance of *E and the value of
// field Code and Name match with values in e.
func (e *E) Is(target error) bool {
	var (
		etarget *E
		ok      bool
	)
	etarget, ok = target.(*E)
	if !ok {
		return false
	}
	return e.Code == etarget.Code && e.Name == etarget.Name
}

// Unwrap return the internal error.
func (e *E) Unwrap() error {
	return e.err
}
