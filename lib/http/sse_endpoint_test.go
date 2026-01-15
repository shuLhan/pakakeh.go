// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package http

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestSSEEndpoint(t *testing.T) {
	var (
		opts = ServerOptions{
			Address: `127.0.0.1:24168`,
		}
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
	var sse = SSEEndpoint{
		Path: `/sse`,
	}

	var err = httpd.RegisterSSE(sse)

	test.Assert(t, `error`, `RegisterSSE: Call field not set`, err.Error())
}

func testSSEEndpointDuplicatePath(t *testing.T, httpd *Server) {
	var ep = Endpoint{
		Path: `/sse`,
		Call: func(_ *EndpointRequest) ([]byte, error) { return nil, nil },
	}

	var err = httpd.RegisterEndpoint(ep)
	if err != nil {
		t.Fatal(err)
	}

	var sse = SSEEndpoint{
		Path: `/sse`,
		Call: func(_ *SSEConn) {},
	}

	err = httpd.RegisterSSE(sse)

	test.Assert(t, `error`, `RegisterSSE: ambigous endpoint`, err.Error())
}
