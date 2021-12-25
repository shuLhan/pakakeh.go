// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

//
// CORSOptions define optional options for server to allow other servers to
// access its resources.
//
type CORSOptions struct {
	exposeHeaders string
	maxAge        string

	// AllowOrigins contains global list of cross-site Origin that are
	// allowed during preflight requests by the OPTIONS method.
	// The list is case-sensitive.
	// To allow all Origin, one must add "*" string to the list.
	AllowOrigins []string

	// AllowHeaders contains global list of allowed headers during
	// preflight requests by the OPTIONS method.
	// The list is case-insensitive.
	// To allow all headers, one must add "*" string to the list.
	AllowHeaders []string

	// ExposeHeaders contains list of allowed headers.
	// This list will be send when browser request OPTIONS without
	// request-method.
	ExposeHeaders []string

	// MaxAge gives the value in seconds for how long the response to
	// the preflight request can be cached for without sending another
	// preflight request.
	MaxAge int

	// AllowCredentials indicates whether or not the actual request
	// can be made using credentials.
	AllowCredentials bool

	allowHeadersAll bool // flag to indicate wildcards on list.
	allowOriginsAll bool // flag to indicate wildcards on list.
}
