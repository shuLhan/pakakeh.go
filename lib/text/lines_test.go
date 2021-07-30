// Copyright 2021 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package text

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseLines(t *testing.T) {
	cases := []struct {
		raw []byte
		exp Lines
	}{{
		raw: []byte(`with single line`),
		exp: Lines{{
			V: []byte(`with single line`),
		}},
	}, {
		raw: []byte(`with
multiple
lines
`),
		exp: Lines{{
			V: []byte(`with`),
		}, {
			N: 1,
			V: []byte(`multiple`),
		}, {
			N: 2,
			V: []byte(`lines`),
		}},
	}}

	for _, c := range cases {
		got := ParseLines(c.raw)
		test.Assert(t, "ParseLines", c.exp, got)
	}
}
