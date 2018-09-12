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

func TestIsFileExist(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc, parent, relpath string
		exp                   bool
	}{{
		desc:    "With directory",
		relpath: "testdata",
	}, {
		desc:    "With non existen path",
		parent:  "/random",
		relpath: "file",
	}, {
		desc:    "With file exist without parent",
		relpath: "testdata/file",
		exp:     true,
	}, {
		desc:    "With file exist",
		parent:  wd,
		relpath: "testdata/file",
		exp:     true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsFileExist(c.parent, c.relpath)

		test.Assert(t, "", c.exp, got, true)
	}
}
