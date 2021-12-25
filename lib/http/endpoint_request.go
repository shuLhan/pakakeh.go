// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

//
// EndpointRequest wrap the called Endpoint and common two parameters in HTTP
// handler: the http.ResponseWriter and http.Request.
//
// The RequestBody field contains the full http.Request.Body that has been
// read.
//
// The Error field is used by CallbackErrorHandler.
//
type EndpointRequest struct {
	HttpWriter  http.ResponseWriter
	Error       error
	Endpoint    *Endpoint
	HttpRequest *http.Request
	RequestBody []byte
}
