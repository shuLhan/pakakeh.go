// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func testRouteHandler(t *testing.T, target string) RouteHandler {
	return func(ctx context.Context, req *Request) (res Response) {
		test.Assert(t, "routeHandler", target, req.Target)
		return
	}
}

func testRootRouteAdd(t *testing.T, defMethod string) {
	type testCase struct {
		expErr  error
		exp     *route
		handler RouteHandler
		desc    string
		method  string
		target  string
	}

	var cases = []testCase{{
		desc:    "With invalid method",
		method:  "PUSH",
		target:  "/",
		handler: testRouteHandler(t, "/"),
		expErr:  ErrRouteInvMethod,
	}, {
		desc:    "Without absolute path",
		method:  defMethod,
		target:  ":id/xyz",
		handler: testRouteHandler(t, ":id/xyz"),
		expErr:  ErrRouteInvTarget,
		exp: &route{
			name: "/",
		},
	}, {
		desc:    "With parameter at first path",
		method:  defMethod,
		target:  "/:id/xyz",
		handler: testRouteHandler(t, "/:id/xyz"),
		exp: &route{
			name: "/",
			childs: []*route{{
				name:    "id",
				isParam: true,
				childs: []*route{{
					name:    "xyz",
					handler: testRouteHandler(t, "/:id/xyz"),
				}},
			}},
		},
	}, {
		desc:    "With duplicate parameter",
		method:  defMethod,
		target:  "/:param/abc",
		expErr:  ErrRouteDupParam,
		handler: testRouteHandler(t, "/:id/xyz"),
		exp: &route{
			name: "/",
			childs: []*route{{
				name:    "id",
				isParam: true,
				childs: []*route{{
					name:    "xyz",
					handler: testRouteHandler(t, "/:id/xyz"),
				}},
			}},
		},
	}, {
		desc:    "With handle on root",
		method:  defMethod,
		target:  "/",
		handler: testRouteHandler(t, "/"),
		exp: &route{
			name:    "/",
			handler: testRouteHandler(t, "/"),
			childs: []*route{{
				name:    "id",
				isParam: true,
				childs: []*route{{
					name:    "xyz",
					handler: testRouteHandler(t, "/:id/xyz"),
				}},
			}},
		},
	}, {
		desc:    "With different sub path",
		method:  defMethod,
		target:  "/:id/abc",
		handler: testRouteHandler(t, "/:id/abc"),
		exp: &route{
			name:    "/",
			handler: testRouteHandler(t, "/"),
			childs: []*route{{
				name:    "id",
				isParam: true,
				childs: []*route{{
					name:    "xyz",
					handler: testRouteHandler(t, "/:id/xyz"),
				}, {
					name:    "abc",
					handler: testRouteHandler(t, "/:id/abc"),
				}},
			}},
		},
	}, {
		desc:    "With another parameter at the end",
		method:  defMethod,
		target:  "/:id/abc/def/:000",
		handler: testRouteHandler(t, "/:id/abc/def/:000"),
		exp: &route{
			name:    "/",
			handler: testRouteHandler(t, "/"),
			childs: []*route{{
				name:    "id",
				isParam: true,
				childs: []*route{{
					name:    "xyz",
					handler: testRouteHandler(t, "/:id/xyz"),
				}, {
					name:    "abc",
					handler: testRouteHandler(t, "/:id/abc"),
					childs: []*route{{
						name: "def",
						childs: []*route{{
							name:    "000",
							isParam: true,
							handler: testRouteHandler(t, "/:id/abc/def/:000"),
						}},
					}},
				}},
			}},
		},
	}}

	var (
		rootRoute = newRootRoute()

		c   testCase
		err error
		got *route
	)
	for _, c = range cases {
		t.Logf("%s: %s %s", c.desc, c.method, c.target)

		err = rootRoute.add(c.method, c.target, c.handler)
		if err != nil {
			test.Assert(t, "err", c.expErr, err)
		}

		got = rootRoute.getParent(c.method)

		test.Assert(t, "route", fmt.Sprintf("%+v", c.exp), fmt.Sprintf("%+v", got))
	}
}

func TestRootRoute_Get(t *testing.T) {
	type testCase struct {
		expParams targetParam
		method    string
		target    string
		expTarget string
	}

	var cases = []testCase{{
		method:    http.MethodGet,
		target:    "/1000/xyz",
		expTarget: "/:id/xyz",
		expParams: targetParam{"id": "1000"},
	}, {
		// Invalid method
		method: "PUSH",
		target: "/1000/xyz",
	}, {
		// Invalid target
		method: http.MethodGet,
		target: "1000/xy",
	}, {
		// Invalid target
		method: http.MethodGet,
		target: "/1000/xy",
	}, {
		method:    http.MethodGet,
		target:    "/",
		expTarget: "/",
		expParams: targetParam{},
	}, {
		method:    http.MethodGet,
		target:    "/333/abc",
		expTarget: "/:id/abc",
		expParams: targetParam{"id": "333"},
	}, {
		method:    http.MethodGet,
		target:    "/333/abc/",
		expTarget: "/:id/abc",
		expParams: targetParam{"id": "333"},
	}, {
		method:    http.MethodGet,
		target:    "/333/abc/def",
		expTarget: "/:id/abc/def",
		expParams: targetParam{"id": "333"},
	}, {
		method: http.MethodGet,
		target: "/333/abc/444",
	}, {
		method: http.MethodGet,
		target: "/333/abc/444/",
	}, {
		method:    http.MethodGet,
		target:    "/333/abc/def/444",
		expTarget: "/:id/abc/def/:000",
		expParams: targetParam{"id": "333", "000": "444"},
	}, {
		method:    http.MethodGet,
		target:    "/333/abc/def/444/",
		expTarget: "/:id/abc/def/:000",
		expParams: targetParam{"id": "333", "000": "444"},
	}, {
		method: http.MethodGet,
		target: "/333/abc/def/444/ghi",
	}}

	var (
		rootRoute = newRootRoute()

		c          testCase
		gotParams  targetParam
		gotHandler RouteHandler
		err        error
	)

	err = rootRoute.add(http.MethodGet, `/:id/xyz`, testRouteHandler(t, `/:id/xyz`))
	if err != nil {
		t.Fatal(err)
	}

	err = rootRoute.add(http.MethodGet, `/:id/abc`, testRouteHandler(t, `/:id/abc`))
	if err != nil {
		t.Fatal(err)
	}

	err = rootRoute.add(http.MethodGet, `/:id/abc/def`, testRouteHandler(t, `/:id/abc/def`))
	if err != nil {
		t.Fatal(err)
	}

	err = rootRoute.add(http.MethodGet, `/:id/abc/def/:000`, testRouteHandler(t, `/:id/abc/def/:000`))
	if err != nil {
		t.Fatal(err)
	}

	for _, c = range cases {
		t.Log(c.method + " " + c.target)

		gotParams, gotHandler = rootRoute.get(c.method, c.target)

		test.Assert(t, "params", c.expParams, gotParams)

		if gotHandler != nil {
			gotHandler(context.Background(), &Request{Target: c.expTarget})
		}
	}
}

func TestRootRoute(t *testing.T) {
	testRootRouteAdd(t, http.MethodDelete)
	testRootRouteAdd(t, http.MethodGet)
	testRootRouteAdd(t, http.MethodPatch)
	testRootRouteAdd(t, http.MethodPost)
	testRootRouteAdd(t, http.MethodPut)
}
