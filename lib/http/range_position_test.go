package http

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseContentRange(t *testing.T) {
	type testCase struct {
		exp *RangePosition
		v   string
	}

	var cases = []testCase{{
		v: ``,
	}, {
		v: `bytes -1`,
	}, {
		v: `bytes -x/10`,
	}, {
		v: `bytes 1`,
	}, {
		v: `bytes x-10/10`,
	}, {
		v: `bytes 10-x/10`,
	}, {
		v: `bytes 10-20/20-`,
	}}

	var (
		c   testCase
		got *RangePosition
	)
	for _, c = range cases {
		got = ParseContentRange(c.v)
		test.Assert(t, c.v, c.exp, got)
	}
}
