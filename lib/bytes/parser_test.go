// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParserRead(t *testing.T) {
	type testCase struct {
		expToken []byte
		expDelim byte
	}

	var (
		parser = NewParser([]byte("a b\tc"), []byte(" \t"))

		cases = []testCase{{
			expToken: []byte(`a`),
			expDelim: ' ',
		}, {
			expToken: []byte(`b`),
			expDelim: '\t',
		}, {
			expToken: []byte(`c`),
			expDelim: 0,
		}, {
			// empty.
		}}

		c     testCase
		token []byte
		d     byte
	)

	for _, c = range cases {
		token, d = parser.Read()
		test.Assert(t, `token`, c.expToken, token)
		test.Assert(t, `delimiter`, c.expDelim, d)
	}
}

func TestParserSkipLine(t *testing.T) {
	type testCase struct {
		expToken []byte
		expDelim byte
	}

	var (
		parser = NewParser([]byte("a\nb\nc\nd e\n"), []byte("\n"))

		cases = []testCase{{
			expToken: []byte(`b`),
			expDelim: '\n',
		}, {
			expToken: []byte(`d e`),
			expDelim: '\n',
		}, {
			// empty.
		}}

		c     testCase
		token []byte
		d     byte
	)

	for _, c = range cases {
		parser.SkipLine()
		token, d = parser.Read()
		test.Assert(t, `token`, c.expToken, token)
		test.Assert(t, `delimiter`, c.expDelim, d)
	}
}
