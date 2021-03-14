// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestRowIsEqual(t *testing.T) {
	cases := []struct {
		a, b Row
		exp  bool
	}{{
		a:   Row{{"a"}, {"b", "c"}},
		b:   Row{{"a"}, {"b", "c"}},
		exp: true,
	}, {
		a:   Row{{"a"}, {"b", "c"}},
		b:   Row{{"a"}, {"c", "b"}},
		exp: true,
	}, {
		a:   Row{{"a"}, {"b", "c"}},
		b:   Row{{"c", "b"}, {"a"}},
		exp: true,
	}, {
		a: Row{{"a"}, {"b", "c"}},
		b: Row{{"a"}, {"b", "a"}},
	}}

	for _, c := range cases {
		got := c.a.IsEqual(c.b)
		test.Assert(t, "", c.exp, got)
	}
}

func TestRowJoin(t *testing.T) {
	cases := []struct {
		row        Row
		lsep, ssep string
		exp        string
	}{{
		//
		lsep: ";",
		ssep: ",",
		exp:  "",
	}, {
		row:  Row{{"a"}, {}},
		lsep: ";",
		ssep: ",",
		exp:  "a;",
	}, {
		row:  Row{{"a"}, {"b", "c"}},
		lsep: ";",
		ssep: ",",
		exp:  "a;b,c",
	}}

	for _, c := range cases {
		got := c.row.Join(c.lsep, c.ssep)
		test.Assert(t, "", c.exp, got)
	}
}
