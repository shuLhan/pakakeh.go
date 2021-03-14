// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/shuLhan/share/lib/memfs"
	"github.com/shuLhan/share/lib/test"
)

var memFS *memfs.MemFS

func TestGeneratePathNode(t *testing.T) {
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

	expExcludeIndexHTML.SetMode(0644)
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

	for _, c := range cases {
		t.Log(c.path)

		got, err := memFS.Get(c.path)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		childs := got.Childs
		got.Childs = nil
		test.Assert(t, "Node", c.exp, got)
		got.Childs = childs
	}
}

func TestNode_Readdir(t *testing.T) {
	cases := []struct {
		path string
		exp  []string
	}{{
		path: "/",
		exp: []string{
			"direct",
			"exclude",
			"include",
			"index.css",
			"index.html",
			"index.js",
			"plain",
		},
	}, {
		path: "/direct",
		exp: []string{
			"add",
		},
	}, {
		path: "/direct/add",
		exp: []string{
			"file",
			"file2",
		},
	}, {
		path: "/exclude",
		exp: []string{
			"index.css",
			"index.html",
			"index.js",
		},
	}, {
		path: "/include",
		exp: []string{
			"index.css",
			"index.html",
			"index.js",
		},
	}}

	for _, c := range cases {
		t.Logf(c.path)

		file, err := memFS.Open(c.path)
		if err != nil {
			t.Fatal(err)
		}

		fis, err := file.Readdir(0)
		if err != nil {
			t.Fatal(err)
		}

		got := make([]string, 0, len(fis))

		for _, fi := range fis {
			got = append(got, fi.Name())
		}

		sort.Strings(got)

		test.Assert(t, "Node.Readdir", c.exp, got)
	}
}
