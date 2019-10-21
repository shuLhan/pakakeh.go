// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// RequestMethod define type of HTTP method.
type RequestMethod int

// List of known HTTP methods.
const (
	RequestMethodGet     RequestMethod = 0
	RequestMethodConnect RequestMethod = 1 << iota
	RequestMethodDelete
	RequestMethodHead
	RequestMethodOptions
	RequestMethodPatch
	RequestMethodPost
	RequestMethodPut
	RequestMethodTrace
)
