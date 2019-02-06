// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestDecodeQP(t *testing.T) {
	cases := []struct {
		in  []byte
		exp []byte
	}{{
		in: []byte{},
	}, {
		in:  []byte("="),
		exp: []byte("="),
	}, {
		in:  []byte("=2"),
		exp: []byte("=2"),
	}, {
		in:  []byte("=20"),
		exp: []byte(" "),
	}, {
		in:  []byte("A=20B"),
		exp: []byte("A B"),
	}, {
		in:  []byte("A\r\n=20B"),
		exp: []byte("A B"),
	}, {
		in:  []byte("A\r\n=2xB"),
		exp: []byte("A=2xB"),
	}}

	for _, c := range cases {
		t.Log(c.in)

		got := DecodeQP(c.in)

		test.Assert(t, "DecodeQP", c.exp, got, true)
	}
}
