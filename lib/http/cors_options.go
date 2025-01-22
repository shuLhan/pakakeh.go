// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package http

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions define optional options for server to allow other servers to
// access its resources.
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

	allowHeadersAll bool // flag to indicate wildcards on AllowHeaders.
	allowOriginsAll bool // flag to indicate wildcards on AllowOrigins.
}

// handle handle the CORS request.
//
// Reference: https://www.html5rocks.com/static/images/cors_server_flowchart.png
func (cors *CORSOptions) handle(res http.ResponseWriter, req *http.Request) {
	var preflightOrigin = req.Header.Get(HeaderOrigin)
	if len(preflightOrigin) == 0 {
		return
	}

	// Set the "Access-Control-Allow-Origin" header based on the request
	// Origin and matched allowed origin.
	// If one of the AllowOrigins contains wildcard "*", then allow all.

	if cors.allowOriginsAll {
		res.Header().Set(HeaderACAllowOrigin, preflightOrigin)
	} else {
		var origin string
		for _, origin = range cors.AllowOrigins {
			if origin == corsWildcard {
				res.Header().Set(HeaderACAllowOrigin, preflightOrigin)
				break
			}
			if origin == preflightOrigin {
				res.Header().Set(HeaderACAllowOrigin, preflightOrigin)
				break
			}
		}
	}

	// Set the "Access-Control-Allow-Method" header based on the request
	// header "Access-Control-Request-Method", only allow HTTP method
	// DELETE, GET, PATCH, POST, and PUT.
	// If no "Access-Control-Request-Method", set the response header
	// "Access-Control-Expose-Headers" based on predefined values.

	var preflightMethod = req.Header.Get(HeaderACRequestMethod)
	if len(preflightMethod) == 0 {
		if len(cors.exposeHeaders) > 0 {
			res.Header().Set(HeaderACExposeHeaders, cors.exposeHeaders)
		}
	} else if preflightMethod == http.MethodDelete ||
		preflightMethod == http.MethodGet ||
		preflightMethod == http.MethodPatch ||
		preflightMethod == http.MethodPost ||
		preflightMethod == http.MethodPut {
		res.Header().Set(HeaderACAllowMethod, preflightMethod)
	}

	cors.handleRequestHeaders(res, req)

	if len(cors.maxAge) > 0 {
		res.Header().Set(HeaderACMaxAge, cors.maxAge)
	}
	if cors.AllowCredentials {
		res.Header().Set(HeaderACAllowCredentials, `true`)
	}
}

// handleRequestHeaders set the response header
// "Access-Control-Allow-Headers" based on the request header
// "Access-Control-Request-Headers".
// If [CORSOptions.AllowHeaders] is empty, no requested headers will be
// allowed.
// If [CORSOptions.AllowHeaders] contains wildcard "*", all requested
// headers are allowed.
func (cors *CORSOptions) handleRequestHeaders(res http.ResponseWriter, req *http.Request) {
	var preflightHeaders = req.Header.Get(HeaderACRequestHeaders)
	if len(preflightHeaders) == 0 {
		return
	}

	var (
		reqHeaders = strings.Split(preflightHeaders, `,`)
		x          int
	)
	for x = range len(reqHeaders) {
		reqHeaders[x] = strings.ToLower(strings.TrimSpace(reqHeaders[x]))
	}

	var (
		allowHeaders = make([]string, 0, len(reqHeaders))
		reqHeader    string
		allowHeader  string
	)
	for _, reqHeader = range reqHeaders {
		if cors.allowHeadersAll {
			allowHeaders = append(allowHeaders, reqHeader)
		} else {
			for _, allowHeader = range cors.AllowHeaders {
				if reqHeader == allowHeader {
					allowHeaders = append(allowHeaders, reqHeader)
					break
				}
			}
		}
	}
	if len(allowHeaders) == 0 {
		return
	}

	res.Header().Set(HeaderACAllowHeaders, strings.Join(allowHeaders, `,`))
}

func (cors *CORSOptions) init() {
	var value string

	for _, value = range cors.AllowOrigins {
		if value == corsWildcard {
			cors.allowOriginsAll = true
			break
		}
	}

	var x int
	for x, value = range cors.AllowHeaders {
		if value == corsWildcard {
			cors.allowHeadersAll = true
		} else {
			cors.AllowHeaders[x] = strings.ToLower(cors.AllowHeaders[x])
		}
	}

	if len(cors.ExposeHeaders) > 0 {
		cors.exposeHeaders = strings.Join(cors.ExposeHeaders, `,`)
	}
	if cors.MaxAge > 0 {
		cors.maxAge = strconv.Itoa(cors.MaxAge)
	}
}
