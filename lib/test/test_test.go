// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"testing"
)

func TestAssert(t *testing.T) {
	cases := []struct {
		in   interface{}
		exp  interface{}
		desc string
	}{
		{
			desc: "With nil",
			in:   nil,
			exp:  nil,
		},
	}

	for _, c := range cases {
		t.Log(c.desc)

		Assert(t, "interface{}", c.exp, c.in)
	}
}
