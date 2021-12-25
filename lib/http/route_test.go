// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewRoute(t *testing.T) {
	cases := []struct {
		desc     string
		ep       *Endpoint
		exp      *route
		expError string
	}{{
		desc: "With empty path",
		ep: &Endpoint{
			Path: "",
		},
		exp: &route{
			path:  "/",
			nodes: []*node{{}},
			endpoint: &Endpoint{
				Path: "",
			},
		},
	}, {
		desc: "With root path",
		ep: &Endpoint{
			Path: "/",
		},
		exp: &route{
			path:  "/",
			nodes: []*node{{}},
			endpoint: &Endpoint{
				Path: "/",
			},
		},
	}, {
		desc: "With empty key",
		ep: &Endpoint{
			Path: "/:user /:",
		},
		expError: ErrEndpointKeyEmpty.Error(),
	}, {
		desc: "With duplicate keys",
		ep: &Endpoint{
			Path: "/:user/a/b/:user/c",
		},
		expError: ErrEndpointKeyDuplicate.Error(),
	}, {
		desc: "With valid keys",
		ep: &Endpoint{
			Path: "/: user/ :repo ",
		},
		exp: &route{
			path: "/:user/:repo",
			nodes: []*node{{
				key:   "user",
				isKey: true,
			}, {
				key:   "repo",
				isKey: true,
			}},
			nkey: 2,
			endpoint: &Endpoint{
				Path: "/: user/ :repo ",
			},
		},
	}, {
		desc: "With double slash on path",
		ep: &Endpoint{
			Path: "/user//repo",
		},
		exp: &route{
			path: "/user/repo",
			nodes: []*node{{
				name: "user",
			}, {
				name: "repo",
			}},
			endpoint: &Endpoint{
				Path: "/user//repo",
			},
		},
	}, {
		desc: "Without key",
		ep: &Endpoint{
			Path: "/user/repo",
		},
		exp: &route{
			path: "/user/repo",
			nodes: []*node{{
				name: "user",
			}, {
				name: "repo",
			}},
			endpoint: &Endpoint{
				Path: "/user/repo",
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := newRoute(c.ep)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		test.Assert(t, "newRoute", c.exp, got)
	}
}

func TestRoute_parse(t *testing.T) {
	type testPath struct {
		expVals map[string]string
		path    string
		expOK   bool
	}

	cases := []struct {
		desc  string
		ep    *Endpoint
		paths []testPath
	}{{
		desc: "With empty path",
		ep: &Endpoint{
			Path: "",
		},
		paths: []testPath{{
			path:  "/",
			expOK: true,
		}, {
			path: "/a",
		}},
	}, {
		desc: "With single key at the beginning",
		ep: &Endpoint{
			Path: "/:user/repo",
		},
		paths: []testPath{{
			path: "/",
		}, {
			path: "/me",
		}, {
			path:  "/me/repo",
			expOK: true,
			expVals: map[string]string{
				"user": "me",
			},
		}, {
			path:  "/me/repo/",
			expOK: true,
			expVals: map[string]string{
				"user": "me",
			},
		}},
	}, {
		desc: "With single key at the middle",
		ep: &Endpoint{
			Path: "/your/:user/repo",
		},
		paths: []testPath{{
			path: "/",
		}, {
			path: "/your",
		}, {
			path: "/your/name",
		}, {
			path:  "/your/name/repo",
			expOK: true,
			expVals: map[string]string{
				"user": "name",
			},
		}, {
			path:  "/your/name/repo/",
			expOK: true,
			expVals: map[string]string{
				"user": "name",
			},
		}, {
			path: "/your/name/repo/here",
		}},
	}, {
		desc: "With single key at the end",
		ep: &Endpoint{
			Path: "/your/user/:repo",
		},
		paths: []testPath{{
			path: "/",
		}, {
			path: "/your",
		}, {
			path: "/your/user",
		}, {
			path:  "/your/user/x",
			expOK: true,
			expVals: map[string]string{
				"repo": "x",
			},
		}, {
			path:  "/your/user/x/",
			expOK: true,
			expVals: map[string]string{
				"repo": "x",
			},
		}, {
			path: "/your/name/x",
		}},
	}, {
		desc: "With double keys",
		ep: &Endpoint{
			Path: "/:user /: repo ",
		},
		paths: []testPath{{
			path: "/",
		}, {
			path: "/user",
		}, {
			path:  "/user/repo",
			expOK: true,
			expVals: map[string]string{
				"user": "user",
				"repo": "repo",
			},
		}, {
			path: "/user/repo/here",
		}},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		rute, err := newRoute(c.ep)
		if err != nil {
			continue
		}

		for _, tp := range c.paths {
			gotVals, gotOK := rute.parse(tp.path)

			test.Assert(t, "vals", tp.expVals, gotVals)
			test.Assert(t, "ok", tp.expOK, gotOK)
		}
	}
}
