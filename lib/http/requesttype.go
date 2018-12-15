// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// RequestType define type of HTTP request.
type RequestType int

// List of valid request type.
const (
	RequestTypeNone  RequestType = 0
	RequestTypeQuery             = 1 << iota
	RequestTypeForm
	RequestTypeMultipartForm
	RequestTypeJSON
)
