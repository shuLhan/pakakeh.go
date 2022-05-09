// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

// RequestMethod define type of HTTP method.
type RequestMethod int

// List of known HTTP methods.
const (
	RequestMethodGet RequestMethod = iota
	RequestMethodConnect
	RequestMethodDelete
	RequestMethodHead
	RequestMethodOptions
	RequestMethodPatch
	RequestMethodPost
	RequestMethodPut
	RequestMethodTrace
)

// String return the string representation of request method.
func (rm RequestMethod) String() string {
	switch rm {
	case RequestMethodGet:
		return http.MethodGet
	case RequestMethodConnect:
		return http.MethodConnect
	case RequestMethodDelete:
		return http.MethodDelete
	case RequestMethodHead:
		return http.MethodHead
	case RequestMethodOptions:
		return http.MethodOptions
	case RequestMethodPatch:
		return http.MethodPatch
	case RequestMethodPost:
		return http.MethodPost
	case RequestMethodPut:
		return http.MethodPut
	case RequestMethodTrace:
		return http.MethodTrace
	}
	return ""
}
