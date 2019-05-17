// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

import (
	"net"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestMacroExpandIPv4(t *testing.T) {
	ref := newResult(net.ParseIP("192.0.2.3"), "email.example.com",
		"strong-bad@email.example.com", "mail.localhost")

	cases := []struct {
		mode   int
		data   string
		exp    string
		expErr string
	}{{
		data: "%{s}",
		exp:  "strong-bad@email.example.com",
	}, {
		data: "%{o}",
		exp:  "email.example.com",
	}, {
		data: "%{d}",
		exp:  "email.example.com",
	}, {
		data: "%{d10}",
		exp:  "email.example.com",
	}, {
		data: "%{d3}",
		exp:  "email.example.com",
	}, {
		data: "%{d2}",
		exp:  "example.com",
	}, {
		data: "%{d1}",
		exp:  "com",
	}, {
		data:   "%{d1",
		exp:    "com",
		expErr: "missing closing '}'",
	}, {
		data:   "%{d1r",
		exp:    "com",
		expErr: "missing closing '}'",
	}, {
		data: "%{d2}.%{d1r}",
		exp:  "example.com.email",
	}, {
		data: "%{dr}",
		exp:  "com.example.email",
	}, {
		data: "%{d2r}",
		exp:  "example.email",
	}, {
		data: "%{l}",
		exp:  "strong-bad",
	}, {
		data: "%{l-}",
		exp:  "strong.bad",
	}, {
		data:   "%{l-",
		expErr: "missing closing '}'",
	}, {
		data:   "%{l-1",
		expErr: "missing closing '}', got '1'",
	}, {
		data: "%{lr}",
		exp:  "strong-bad",
	}, {
		data: "%{lr-}",
		exp:  "bad.strong",
	}, {
		data: "%{l1r-}",
		exp:  "strong",
	}, {
		data: "%{ir}.%{v}._spf.%{d2}",
		exp:  "3.2.0.192.in-addr._spf.example.com",
	}, {
		data: "%{lr-}.lp._spf.%{d2}",
		exp:  "bad.strong.lp._spf.example.com",
	}, {
		data: "%{lr-}.lp.%{ir}.%{v}._spf.%{d2}",
		exp:  "bad.strong.lp.3.2.0.192.in-addr._spf.example.com",
	}, {
		data: "%{ir}.%{v}.%{l1r-}.lp._spf.%{d2}",
		exp:  "3.2.0.192.in-addr.strong.lp._spf.example.com",
	}, {
		data: "%{d2}.trusted-domains.example.net",
		exp:  "example.com.trusted-domains.example.net",
	}, {
		data: "%{hr}",
		exp:  "com.example.email",
	}, {
		mode: modifierExp,
		data: "%{r1r}",
		exp:  "mail",
	}, {
		mode: modifierExp,
		data: "%{c}",
		exp:  "192.0.2.3",
	}, {
		data:   "%%%_%- ",
		expErr: "invalid macro literal ' ' at position 6",
	}, {
		data:   "%%%_%-%d",
		expErr: "syntax error 'd' at position 7",
	}, {
		data:   "%{4r}",
		expErr: "unknown macro letter '4' at position 2",
	}}

	for _, c := range cases {
		t.Log(c.data)

		got, err := macroExpand(ref, c.mode, []byte(c.data))
		if err != nil {
			test.Assert(t, "error", string(c.expErr), err.Error(), true)
			continue
		}

		test.Assert(t, "macroExpand", c.exp, string(got), true)
	}
}

func TestMacroExpandIPv6(t *testing.T) {
	ref := newResult(net.ParseIP("2001:db8::cb01"), "email.example.com",
		"strong-bad@email.example.com", "email.localhost")

	cases := []struct {
		mode   int
		data   string
		exp    string
		expErr string
	}{{
		data: "%{ir}.%{v}._spf.%{d2}",
		exp:  "1.0.b.c.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6._spf.example.com",
	}}

	for _, c := range cases {
		t.Log(c.data)

		got, err := macroExpand(ref, c.mode, []byte(c.data))
		if err != nil {
			test.Assert(t, "error", string(c.expErr), string(got), true)
			continue
		}

		test.Assert(t, "macroExpand", c.exp, string(got), true)
	}
}
