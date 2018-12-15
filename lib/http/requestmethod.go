// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// RequestMethod define type of HTTP method.
type RequestMethod int

const (
	RequestMethodConnect RequestMethod = 1 << iota
	RequestMethodDelete
	RequestMethodGet
	RequestMethodHead
	RequestMethodOptions
	RequestMethodPatch
	RequestMethodPost
	RequestMethodPut
	RequestMethodTrace
)
