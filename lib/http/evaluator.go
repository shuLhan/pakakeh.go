// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

//
// Evaluator is an interface for middleware between actual request and handler.
//
type Evaluator interface {
	//
	// Evaluate the request.  If request is invalid, the error will tell
	// the response code and the error message to be written back to
	// client.
	//
	Evaluate(req *http.Request, reqBody []byte) error
}
