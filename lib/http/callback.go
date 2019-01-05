// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

//
// Callback define a type of function for handling registered handler.
//
// The function will have the query URL, request multipart form data,
// and request body ready to be used in req parameter.
//
// The ResponseWriter can be used to write custom header or to write cookies
// but should not be used to write response body.
//
// The error return type should be instance of StatusError. If error is not
// nil and not *StatusError, server will response with internal-server-error
// status code.
//
type Callback func(res http.ResponseWriter, req *http.Request, reqBody []byte) (resBody []byte, err error)
