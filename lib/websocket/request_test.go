// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestRequestReset(t *testing.T) {
	var req = _reqPool.Get().(*Request)

	req.reset()

	test.Assert(t, "Request.ID", uint64(0), req.ID)
	test.Assert(t, "Request.Method", "", req.Method)
	test.Assert(t, "Request.Target", "", req.Target)
	test.Assert(t, "Request.Body", "", req.Body)
	test.Assert(t, "Request.Path", "", req.Path)
	test.Assert(t, "Request.Params", targetParam{}, req.Params)
	test.Assert(t, "Request.Query", url.Values{}, req.Query)
}

func testHandleGet(_ context.Context, _ *Request) (res Response) {
	return
}

func TestRequestUnpack(t *testing.T) {
	type testCase struct {
		desc      string
		req       *Request
		expParams targetParam
		expQuery  url.Values
		expErr    string
	}

	var (
		rootRoutes = newRootRoute()
	)

	_ = rootRoutes.add(http.MethodGet, "/get/:id", testHandleGet)

	var cases = []testCase{{
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
		expErr: `unpack: invalid URL escape "%in"`,

		// In go version >= 1.20 it will not return an error anymore.
		// See https://github.com/golang/go/issues/56732
		expParams: targetParam{
			"id": "1",
		},
		expQuery: url.Values{
			`%in`: []string{`va`},
		},
	}, {
		desc: "With param",
		req: &Request{
			Method: http.MethodGet,
			Target: "/get/1",
		},
		expParams: targetParam{
			"id": "1",
		},
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
	}}

	var (
		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)

		_, err = c.req.unpack(rootRoutes)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Params", c.expParams, c.req.Params)
		test.Assert(t, "Query", c.expQuery, c.req.Query)
	}
}
