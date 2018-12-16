// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

//
// StatusError define custom error type returned by callback.
//
type StatusError struct {
	Code    int
	Message string
}

//
// Error implement the error interface.
//
func (se *StatusError) Error() string {
	return se.Message
}
