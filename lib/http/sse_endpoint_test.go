// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestSSEEndpoint(t *testing.T) {
	var opts = &ServerOptions{
		Address: `127.0.0.1:24168`,
	}

	var (
		httpd *Server
		err   error
	)
	httpd, err = NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}

	t.Run(`EmptyCall`, func(tt *testing.T) {
		testSSEEndpointEmptyCall(tt, httpd)
	})

	t.Run(`DuplicatePath`, func(tt *testing.T) {
		testSSEEndpointDuplicatePath(tt, httpd)
	})
}

func testSSEEndpointEmptyCall(t *testing.T, httpd *Server) {
	var sse = &SSEEndpoint{
		Path: `/sse`,
	}

	var err = httpd.RegisterSSE(sse)

	test.Assert(t, `error`, `RegisterSSE: Call field not set`, err.Error())
}

func testSSEEndpointDuplicatePath(t *testing.T, httpd *Server) {
	var ep = &Endpoint{
		Path: `/sse`,
		Call: func(_ *EndpointRequest) ([]byte, error) { return nil, nil },
	}

	var err = httpd.RegisterEndpoint(ep)
	if err != nil {
		t.Fatal(err)
	}

	var sse = &SSEEndpoint{
		Path: `/sse`,
		Call: func(_ *SSEConn) {},
	}

	err = httpd.RegisterSSE(sse)

	test.Assert(t, `error`, `RegisterSSE: ambigous endpoint`, err.Error())
}
