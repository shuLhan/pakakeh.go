// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

// RequestMethod define type of HTTP method.
type RequestMethod string

// List of known HTTP methods.
const (
	RequestMethodConnect RequestMethod = http.MethodConnect
	RequestMethodDelete  RequestMethod = http.MethodDelete
	RequestMethodGet     RequestMethod = http.MethodGet
	RequestMethodHead    RequestMethod = http.MethodHead
	RequestMethodOptions RequestMethod = http.MethodOptions
	RequestMethodPatch   RequestMethod = http.MethodPatch
	RequestMethodPost    RequestMethod = http.MethodPost
	RequestMethodPut     RequestMethod = http.MethodPut
	RequestMethodTrace   RequestMethod = http.MethodTrace
)
