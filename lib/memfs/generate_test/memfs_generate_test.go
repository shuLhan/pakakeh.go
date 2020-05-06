// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/memfs"
	"github.com/shuLhan/share/lib/test"
)

func TestGeneratePathNode(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	wd = strings.TrimSuffix(wd, "generate_test")

	expRoot := &memfs.Node{
		SysPath: "testdata",
		Path:    "/",
	}
	expRoot.SetMode(2147484141)
	expRoot.SetName("/")
	expRoot.SetSize(0)

	expExcludeIndexHTML := &memfs.Node{
		SysPath:     filepath.Join("testdata", "exclude", "index.html"),
		Path:        "/exclude/index.html",
		ContentType: "text/html; charset=utf-8",
		V:           []byte("<html></html>\n"),
	}

	expExcludeIndexHTML.SetMode(420) //nolint: staticcheck
	expExcludeIndexHTML.SetName("index.html")
	expExcludeIndexHTML.SetSize(14)

	cases := []struct {
		path     string
		exp      *memfs.Node
		expError string
	}{{
		path:     "/gen_test.go",
		expError: "file does not exist",
	}, {
		path: "/",
		exp:  expRoot,
	}, {
		path: "/exclude/index.html",
		exp:  expExcludeIndexHTML,
	}}

	mfs, err := memfs.New("", nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c.path)

		got, err := mfs.Get(c.path)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		got.SysPath = strings.Replace(got.SysPath, "xxx/", wd, -1)

		test.Assert(t, "Node", c.exp, got, true)
	}
}
