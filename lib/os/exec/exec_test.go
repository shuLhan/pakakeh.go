// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package exec

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
	}, {
		in:      `a\ b c\ d`,
		expCmd:  "a b",
		expArgs: []string{"c d"},
	}, {
		in:      `a\\ b c\\ d`,
		expCmd:  `a\`,
		expArgs: []string{"b", `c\`, "d"},
	}, {
		in:      `a\\\ b c\\\ d`,
		expCmd:  `a\ b`,
		expArgs: []string{`c\ d`},
	}, {
		in:      `sh -c "echo \"a\""`,
		expCmd:  "sh",
		expArgs: []string{`-c`, `echo "a"`},
	}, {
		in:      `sh -c "sh -c \"echo 'a\x'\""`,
		expCmd:  "sh",
		expArgs: []string{`-c`, `sh -c "echo 'a\x'"`},
	}, {
		in:      `sh -c "sh -c \"echo 'a'\'''\""`,
		expCmd:  "sh",
		expArgs: []string{`-c`, `sh -c "echo 'a'\'''"`},
	}, {
		in:      `sh -c "sh -c \"echo 'a\\\"'\""`,
		expCmd:  "sh",
		expArgs: []string{`-c`, `sh -c "echo 'a\"'"`},
	}}

	for _, c := range cases {
		t.Log(c.in)
		gotCmd, gotArgs := ParseCommandArgs(c.in)
		test.Assert(t, "cmd", c.expCmd, gotCmd)
		test.Assert(t, "args", c.expArgs, gotArgs)
	}
}
