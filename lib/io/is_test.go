// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsDirEmpty(t *testing.T) {
	emptyDir := "testdata/dirempty"
	err := os.MkdirAll(emptyDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc string
		path string
		exp  bool
	}{{
		desc: `With dir not exist`,
		path: `testdata/notexist`,
		exp:  true,
	}, {
		desc: `With dir exist and not empty`,
		path: `testdata`,
	}, {
		desc: `With dir exist and empty`,
		path: `testdata/dirempty`,
		exp:  true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsDirEmpty(c.path)

		test.Assert(t, "", c.exp, got, true)
	}
}
