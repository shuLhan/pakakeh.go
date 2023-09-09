// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestConfirmYesNo(t *testing.T) {
	type testCase struct {
		answer   string
		defIsYes bool
		exp      bool
	}

	var cases = []testCase{{
		defIsYes: true,
		exp:      true,
	}, {
		defIsYes: true,
		answer:   "  ",
		exp:      true,
	}, {
		defIsYes: true,
		answer:   "  no",
		exp:      false,
	}, {
		defIsYes: true,
		answer:   " yes",
		exp:      true,
	}, {
		defIsYes: true,
		answer:   " Ys",
		exp:      true,
	}, {
		defIsYes: false,
		exp:      false,
	}, {
		defIsYes: false,
		answer:   "",
		exp:      false,
	}, {

		defIsYes: false,
		answer:   "  no",
		exp:      false,
	}, {
		defIsYes: false,
		answer:   "  yes",
		exp:      true,
	}}

	var (
		mockReader bytes.Buffer
		c          testCase
		got        bool
	)

	for _, c = range cases {
		t.Log(c)
		mockReader.Reset()

		// Write the answer to be read.
		mockReader.WriteString(c.answer + "\n")

		got = ConfirmYesNo(&mockReader, "confirm", c.defIsYes)

		test.Assert(t, "answer", c.exp, got)
	}
}
