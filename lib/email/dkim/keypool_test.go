// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestKeyPoolClear(t *testing.T) {
	DefaultKeyPool.Clear()

	DefaultKeyPool.Put("example.com", &Key{ExpiredAt: 1})
	got := DefaultKeyPool.String()
	test.Assert(t, "DefaultKeyPool.Clear", "[{example.com 1}]", got)

	DefaultKeyPool.Clear()

	got = DefaultKeyPool.String()
	test.Assert(t, "DefaultKeyPool.Clear", "[]", got)
}

func TestKeyPoolPut(t *testing.T) {
	cases := []struct {
		dname string
		key   *Key
		exp   string
	}{{
		dname: "",
		exp:   "[]",
	}, {
		dname: "emptykey",
		exp:   "[]",
	}, {
		dname: "example.com",
		key: &Key{
			Public:    []byte("example.com"),
			ExpiredAt: 1,
		},
		exp: "[{example.com 1}]",
	}, {
		dname: "example.com",
		key: &Key{
			Public:    []byte("example.com"),
			ExpiredAt: 1577811600, // 2020-01-01
		},
		exp: "[{example.com 1577811600}]",
	}, {
		dname: "example.net",
		key: &Key{
			Public:    []byte("example.net"),
			ExpiredAt: 1577811600, // 2020-01-01
		},
		exp: "[{example.com 1577811600}{example.net 1577811600}]",
	}}

	for _, c := range cases {
		t.Log(c.dname)

		DefaultKeyPool.Put(c.dname, c.key)
		got := DefaultKeyPool.String()

		test.Assert(t, "DefaultKeyPool", c.exp, got)
	}
}

func TestKeyPoolGet(t *testing.T) {
	t.Skip("TODO: use local DNS")
	cases := []struct {
		dname  string
		exp    string
		expErr string
	}{{
		dname: "",
	}, {
		dname: "example.com",
		exp:   "v=spf1 -all",
	}, {
		dname: "example.net",
		exp:   "v=spf1 -all",
	}, {
		dname:  "amazon.com",
		expErr: "dkim: LookupKey: multiple TXT records on 'amazon.com'",
	}}

	for _, c := range cases {
		t.Log(c.dname)

		key, err := DefaultKeyPool.Get(c.dname)
		if err != nil {
			serr := err.Error()
			if strings.Contains(serr, "timeout") {
				continue
			}
			test.Assert(t, "error", c.expErr, serr)
			continue
		}
		if key == nil {
			continue
		}

		got := key.Pack()
		test.Assert(t, "DefaultKeyPool.Get", c.exp, got)
	}
}
