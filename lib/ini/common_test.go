// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsValueBoolTrue(t *testing.T) {
	cases := []struct {
		desc string
		v    string
		exp  bool
	}{{
		desc: "With empty value",
	}, {
		desc: "With value in all caps",
		v:    "TRUE",
		exp:  true,
	}, {
		desc: "With value is yes",
		v:    "YES",
		exp:  true,
	}, {
		desc: "With value is ya",
		v:    "yA",
		exp:  true,
	}, {
		desc: "With value is 1",
		v:    "1",
		exp:  true,
	}, {
		desc: "With value is 11",
		v:    "11",
		exp:  false,
	}, {
		desc: "With value is tru (typo)",
		v:    "tru",
		exp:  false,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsValueBoolTrue(c.v)

		test.Assert(t, "", c.exp, got, true)
	}
}
