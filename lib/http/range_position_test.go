package http

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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

func ptrInt64(v int64) *int64 { return &v }

func TestRangePositionContentRange(t *testing.T) {
	var (
		unit = AcceptRangesBytes
		pos  = RangePosition{
			start: ptrInt64(10),
			end:   ptrInt64(20),
		}
	)

	test.Assert(t, ``, `bytes 10-20/512`, pos.ContentRange(unit, 512))
	test.Assert(t, ``, `bytes 10-20/*`, pos.ContentRange(unit, 0))
}
