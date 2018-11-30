// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"net/url"
	"strings"
	"sync"
)

var ( // nolint: gochecknoglobals
	_reqPool = sync.Pool{
		New: func() interface{} {
			return new(Request)
		},
	}
)

//
// Request define text payload format for client requesting resource on
// server.
//
// Example of request format,
//
// 	{
// 		"id": 1512459721269,
//		"method": "GET",
// 		"target": "/v1/auth/login",
// 		"body": "{ \"token\": \"xxx.yyy.zzz\" }"
// 	}
//
type Request struct {
	//
	// Id is unique between request to differentiate multiple request
	// since each request is asynchronous.  Client can use incremental
	// value or, the recommended way, using Unix timestamp with
	// millisecond.
	//
	ID uint64 `json:"id"`

	// Method is equal to HTTP method.
	Method string `json:"method"`

	// Target is equal to HTTP request RequestURI, e.g. "/path?query".
	Target string `json:"target"`

	// Body is equal to HTTP body on POST/PUT.
	Body string `json:"body"`

	// Path is Target without query.
	Path string `json:"-"`

	// Params are parameters as key-value in Target path that has been
	// parsed.
	Params targetParam `json:"-"`

	// Query is Target query.
	Query url.Values `json:"-"`
}

//
// Reset all field's value to zero or empty.
//
func (req *Request) Reset() {
	req.ID = 0
	req.Method = ""
	req.Target = ""
	req.Body = ""
	req.Path = ""
	req.Params = nil
	req.Query = nil
}

//
// unpack the request, parse parameters and query from target.
//
func (req *Request) unpack(routes *rootRoute) (handler RouteHandler, err error) {
	pathQuery := strings.SplitN(req.Target, pathQuerySep, 2)
	if len(pathQuery) == 0 {
		return
	}

	req.Path = pathQuery[0]

	req.Params, handler = routes.get(req.Method, req.Target)
	if handler == nil {
		return
	}

	if len(pathQuery) == 2 {
		req.Query, err = url.ParseQuery(pathQuery[1])
		if err != nil {
			return
		}
	}

	return
}
