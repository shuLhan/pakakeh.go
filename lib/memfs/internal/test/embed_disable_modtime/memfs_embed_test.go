// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package embed

import (
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/memfs"
	"github.com/shuLhan/share/lib/test"
)

var memFS *memfs.MemFS

func TestGeneratePathNode(t *testing.T) {
	zeroTime := time.Time{}

	expRoot := &memfs.Node{
		SysPath:     "testdata",
		Path:        "/",
		GenFuncName: "generate_testdata",
	}
	expRoot.SetMode(2147484141)
	expRoot.SetName("/")
	expRoot.SetSize(0)

	expExcludeIndexHTML := &memfs.Node{
		SysPath:     filepath.Join("testdata", "exclude", "index-link.html"),
		Path:        "/exclude/index-link.html",
		ContentType: "text/html; charset=utf-8",
		Content:     []byte("<html></html>\n"),
		GenFuncName: "generate_testdata_exclude_index_link_html",
	}

	expExcludeIndexHTML.SetMode(0644)
	expExcludeIndexHTML.SetName("index-link.html")
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
		path: "/exclude/index-link.html",
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
		got.SetModTime(zeroTime)
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
			"dir",
			"index-link.css",
			"index-link.html",
			"index-link.js",
		},
	}, {
		path: "/include",
		exp: []string{
			"dir",
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
