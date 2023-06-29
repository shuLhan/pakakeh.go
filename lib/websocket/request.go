// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var (
	_reqPool = sync.Pool{
		New: func() interface{} {
			return new(Request)
		},
	}
)

// Request define text payload format for client requesting resource on
// server.
//
// Example of request format,
//
//	{
//		"id": 1512459721269,
//		"method": "GET",
//		"target": "/v1/auth/login",
//		"body": "{ \"token\": \"xxx.yyy.zzz\" }"
//	}
type Request struct {
	// Query is Target query.
	Query url.Values `json:"-"`

	// Params are parameters as key-value in Target path that has been
	// parsed.
	Params targetParam `json:"-"`

	// Method is equal to HTTP method.
	Method string `json:"method"`

	// Target is equal to HTTP request RequestURI, e.g. "/path?query".
	Target string `json:"target"`

	// Body is equal to HTTP body on POST/PUT.
	Body string `json:"body"`

	// Path is Target without query.
	Path string `json:"-"`

	//
	// Id is unique between request to differentiate multiple request
	// since each request is asynchronous.  Client can use incremental
	// value or, the recommended way, using Unix timestamp with
	// millisecond.
	//
	ID uint64 `json:"id"`

	// Conn is the client connection, where the request come from.
	Conn int
}

// reset all Request field's value to zero.
func (req *Request) reset() {
	req.ID = 0
	req.Method = ""
	req.Target = ""
	req.Body = ""
	req.Path = ""
	req.Params = make(targetParam)
	req.Query = make(url.Values)
}

// unpack the request, parse parameters and query from target.
func (req *Request) unpack(routes *rootRoute) (handler RouteHandler, err error) {
	var logp = `unpack`

	if len(req.Target) == 0 {
		return nil, nil
	}

	var pathQuery []string = strings.SplitN(req.Target, pathQuerySep, 2)

	req.Path = pathQuery[0]

	req.Params, handler = routes.get(req.Method, req.Path)
	if handler == nil {
		return nil, nil
	}

	if len(pathQuery) == 2 {
		req.Query, err = url.ParseQuery(pathQuery[1])
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return handler, nil
}
