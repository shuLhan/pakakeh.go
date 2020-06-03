// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseCommandArg(t *testing.T) {
	cases := []struct {
		in      string
		expCmd  string
		expArgs []string
	}{{
		in: ``,
	}, {
		in:      `a `,
		expCmd:  `a`,
		expArgs: nil,
	}, {
		in:      `a "b c"`,
		expCmd:  `a`,
		expArgs: []string{`b c`},
	}, {
		in:      `a "b'c"`,
		expCmd:  `a`,
		expArgs: []string{`b'c`},
	}, {
		in:      `'a "b'c"`,
		expCmd:  `a "b`,
		expArgs: []string{`c`},
	}, {
		in:      "a `b c`",
		expCmd:  `a`,
		expArgs: []string{`b c`},
	}, {
		in:      "a `b'c`",
		expCmd:  `a`,
		expArgs: []string{`b'c`},
	}}

	for _, c := range cases {
		t.Logf(c.in)
		gotCmd, gotArgs := ParseCommandArgs(c.in)
		test.Assert(t, "cmd", c.expCmd, gotCmd, true)
		test.Assert(t, "args", c.expArgs, gotArgs, true)
	}
}
