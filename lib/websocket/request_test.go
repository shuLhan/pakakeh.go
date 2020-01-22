// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestRequestReset(t *testing.T) {
	req := _reqPool.Get().(*Request)

	req.reset()

	test.Assert(t, "Request.ID", uint64(0), req.ID, true)
	test.Assert(t, "Request.Method", "", req.Method, true)
	test.Assert(t, "Request.Target", "", req.Target, true)
	test.Assert(t, "Request.Body", "", req.Body, true)
	test.Assert(t, "Request.Path", "", req.Path, true)
	test.Assert(t, "Request.Params", targetParam{}, req.Params, true)
	test.Assert(t, "Request.Query", url.Values{}, req.Query, true)
}

func testHandleGet(ctx context.Context, req *Request) (res Response) {
	return
}

func TestRequestUnpack(t *testing.T) {
	rootRoutes := newRootRoute()
	_ = rootRoutes.add(http.MethodGet, "/get/:id", testHandleGet)

	cases := []struct {
		desc       string
		req        *Request
		expParams  targetParam
		expQuery   url.Values
		expHandler RouteHandler
		expErr     string
	}{{
		desc: "With empty Target",
		req:  &Request{},
	}, {
		desc: "With empty method",
		req: &Request{
			Target: "/get",
		},
	}, {
		desc: "With unknown target",
		req: &Request{
			Method: http.MethodGet,
			Target: "/unknown",
		},
	}, {
		desc: "With invalid query",
		req: &Request{
			Method: http.MethodGet,
			Target: "/get/1?%in=va",
		},
		expErr: `websocket: Request.unpack: invalid URL escape "%in"`,
	}, {
		desc: "With param",
		req: &Request{
			Method: http.MethodGet,
			Target: "/get/1",
		},
		expParams: targetParam{
			"id": "1",
		},
		expHandler: testHandleGet,
	}, {
		desc: "With query",
		req: &Request{
			Method: http.MethodGet,
			Target: "/get/1?q=2",
		},
		expParams: targetParam{
			"id": "1",
		},
		expQuery: url.Values{
			"q": []string{"2"},
		},
		expHandler: testHandleGet,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, err := c.req.unpack(rootRoutes)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Params", c.expParams, c.req.Params, true)
		test.Assert(t, "Query", c.expQuery, c.req.Query, true)
	}
}
