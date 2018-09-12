// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
	"github.com/shuLhan/share/lib/test/mock"
)

func TestConfirmYesNo(t *testing.T) {
	cases := []struct {
		defIsYes bool
		answer   string
		exp      bool
	}{{
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

	in := mock.Stdin()

	for _, c := range cases {
		t.Log(c)

		in.WriteString(c.answer + "\n")

		mock.ResetStdin(false)

		got := ConfirmYesNo(in, "confirm", c.defIsYes)

		test.Assert(t, "answer", c.exp, got, true)

		mock.ResetStdin(true)
	}
}
