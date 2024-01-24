// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewRoute(t *testing.T) {
	type testCase struct {
		desc     string
		path     string
		exp      *Route
		expError string
	}

	var cases = []testCase{{
		desc: `With empty path`,
		path: ``,
		exp: &Route{
			path:  `/`,
			nodes: []*routeNode{{}},
		},
	}, {
		desc: `With root path`,
		path: `/`,
		exp: &Route{
			path:  `/`,
			nodes: []*routeNode{{}},
		},
	}, {
		desc:     `With empty key`,
		path:     `/:user /:`,
		expError: ErrPathKeyEmpty.Error(),
	}, {
		desc:     `With duplicate keys`,
		path:     `/:user/a/b/:user/c`,
		expError: ErrPathKeyDuplicate.Error(),
	}, {
		desc: `With valid keys`,
		path: `/: user/ :repo `,
		exp: &Route{
			path: `/:user/:repo`,
			nodes: []*routeNode{{
				key:   `user`,
				isKey: true,
			}, {
				key:   `repo`,
				isKey: true,
			}},
			nkey: 2,
		},
	}, {
		desc: `With double slash on path`,
		path: `/user//repo`,
		exp: &Route{
			path: `/user/repo`,
			nodes: []*routeNode{{
				name: `user`,
			}, {
				name: `repo`,
			}},
		},
	}, {
		desc: `Without key`,
		path: `/user/repo`,
		exp: &Route{
			path: `/user/repo`,
			nodes: []*routeNode{{
				name: `user`,
			}, {
				name: `repo`,
			}},
		},
	}}

	var (
		c   testCase
		got *Route
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)

		got, err = NewRoute(c.path)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}

		test.Assert(t, `NewRoute`, c.exp, got)
	}
}

func TestRouteParse(t *testing.T) {
	type testPath struct {
		expVals map[string]string
		path    string
		expOK   bool
	}
	type testCase struct {
		desc  string
		path  string
		paths []testPath
	}

	var cases = []testCase{{
		desc: `With empty path`,
		path: ``,
		paths: []testPath{{
			path:  `/`,
			expOK: true,
		}, {
			path: `/a`,
		}},
	}, {
		desc: `With single key at the beginning`,
		path: `/:user/repo`,
		paths: []testPath{{
			path: `/`,
		}, {
			path: `/me`,
		}, {
			path:  `/me/repo`,
			expOK: true,
			expVals: map[string]string{
				`user`: `me`,
			},
		}, {
			path:  `/me/repo/`,
			expOK: true,
			expVals: map[string]string{
				`user`: `me`,
			},
		}},
	}, {
		desc: `With single key at the middle`,
		path: `/your/:user/repo`,
		paths: []testPath{{
			path: `/`,
		}, {
			path: `/your`,
		}, {
			path: `/your/name`,
		}, {
			path:  `/your/name/repo`,
			expOK: true,
			expVals: map[string]string{
				`user`: `name`,
			},
		}, {
			path:  `/your/name/repo/`,
			expOK: true,
			expVals: map[string]string{
				`user`: `name`,
			},
		}, {
			path: `/your/name/repo/here`,
		}},
	}, {
		desc: `With single key at the end`,
		path: `/your/user/:repo`,
		paths: []testPath{{
			path: `/`,
		}, {
			path: `/your`,
		}, {
			path: `/your/user`,
		}, {
			path:  `/your/user/x`,
			expOK: true,
			expVals: map[string]string{
				`repo`: `x`,
			},
		}, {
			path:  `/your/user/x/`,
			expOK: true,
			expVals: map[string]string{
				`repo`: `x`,
			},
		}, {
			path: `/your/name/x`,
		}},
	}, {
		desc: `With double keys`,
		path: `/:user /: repo `,
		paths: []testPath{{
			path: `/`,
		}, {
			path: `/user`,
		}, {
			path:  `/user/repo`,
			expOK: true,
			expVals: map[string]string{
				`user`: `user`,
				`repo`: `repo`,
			},
		}, {
			path: `/user/repo/here`,
		}},
	}}

	var (
		c       testCase
		rute    *Route
		tp      testPath
		err     error
		gotVals map[string]string
		gotOK   bool
	)
	for _, c = range cases {
		t.Log(c.desc)

		rute, err = NewRoute(c.path)
		if err != nil {
			continue
		}

		for _, tp = range c.paths {
			gotVals, gotOK = rute.Parse(tp.path)

			test.Assert(t, `vals`, tp.expVals, gotVals)
			test.Assert(t, `ok`, tp.expOK, gotOK)
		}
	}
}
