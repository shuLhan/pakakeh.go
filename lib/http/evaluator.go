// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

// Evaluator evaluate the request.  If request is invalid, the error will tell
// the response code and the error message to be written back to client.
type Evaluator func(req *http.Request, reqBody []byte) error
