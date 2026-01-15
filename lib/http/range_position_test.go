// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package http

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseContentRange(t *testing.T) {
	type testCase struct {
		exp      *RangePosition
		expError string
		v        string
	}

	var cases = []testCase{{
		v:        ``,
		expError: `ParseContentRange: invalid Content-Range ""`,
	}, {
		v:        `bytes -1`,
		expError: `ParseContentRange: invalid Content-Range "bytes -1"`,
	}, {
		v:        `bytes -x/10`,
		expError: `ParseContentRange: invalid Content-Range "bytes -x/10"`,
	}, {
		v:        `bytes 1`,
		expError: `ParseContentRange: invalid Content-Range "bytes 1"`,
	}, {
		v:        `bytes x-10/10`,
		expError: `ParseContentRange: invalid Content-Range "bytes x-10/10": strconv.ParseInt: parsing "x": invalid syntax`,
	}, {
		v:        `bytes 10-x/10`,
		expError: `ParseContentRange: invalid Content-Range "bytes 10-x/10": strconv.ParseInt: parsing "x": invalid syntax`,
	}, {
		v:        `bytes 10-20/20-`,
		expError: `ParseContentRange: invalid Content-Range "bytes 10-20/20-"`,
	}}

	var (
		c   testCase
		got *RangePosition
		err error
	)
	for _, c = range cases {
		got, err = ParseContentRange(c.v)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}
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
