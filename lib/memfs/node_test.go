// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNode_Read(t *testing.T) {
	node := &Node{
		size:    3,
		Content: []byte("123"),
	}

	p := make([]byte, 1)

	cases := []struct {
		desc     string
		expError string
		p        []byte
		exp      []byte
		expN     int
	}{{
		desc: "With empty p",
	}, {
		desc: "With buffer 1 byte (1)",
		p:    p,
		exp:  []byte(`1`),
		expN: 1,
	}, {
		desc: "With buffer 1 byte (2)",
		p:    p,
		exp:  []byte(`2`),
		expN: 1,
	}, {
		desc: "With buffer 1 byte (3)",
		p:    p,
		exp:  []byte(`3`),
		expN: 1,
	}, {
		desc:     "With buffer 1 byte (4)",
		p:        p,
		exp:      []byte(`3`),
		expN:     0,
		expError: "Read: EOF",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		n, err := node.Read(c.p)

		test.Assert(t, "p", c.exp, c.p)
		test.Assert(t, "n", c.expN, n)

		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
		}
	}
}

func TestNode_Readdir(t *testing.T) {
	opts := &Options{
		Root: "testdata",
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	f, err := mfs.Open("/")
	if err != nil {
		t.Fatal(err)
	}

	expFileNames := []string{
		"direct",
		"exclude",
		"include",
		"index.css",
		"index.html",
		"index.js",
		"plain",
	}

	fis, err := f.Readdir(0)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Readdir(0)", expFileNames, gotFileNames(fis))

	// Test reading two nodes at a time.

	allFis := make([]os.FileInfo, 0, len(expFileNames))
	for {
		fis, _ := f.Readdir(2)
		if fis == nil {
			break
		}
		allFis = append(allFis, fis...)
	}

	test.Assert(t, "Readdir(2)", expFileNames, gotFileNames(allFis))
}

func gotFileNames(fis []os.FileInfo) (names []string) {
	for _, fi := range fis {
		names = append(names, fi.Name())
	}
	sort.Strings(names)
	return
}

func TestNode_Save(t *testing.T) {
	logp := "TestNode_Save"

	tempDir, err := os.MkdirTemp("", logp)
	if err != nil {
		t.Fatalf("%s: %s", logp, err)
	}

	node := &Node{
		SysPath: filepath.Join(tempDir, "node_save"),
		mode:    0600,
	}

	err = os.WriteFile(node.SysPath, []byte{}, node.mode.Perm())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	cases := []struct {
		desc            string
		content         []byte
		contentEncoding string
		expContent      []byte
	}{{
		desc:       "With non empty content",
		content:    []byte("ABC"),
		expContent: []byte("ABC"),
	}}

	for _, c := range cases {
		err = node.Save(c.content)
		if err != nil {
			t.Fatal(err)
		}

		got, err := os.ReadFile(node.SysPath)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc+": file content", c.content, got)
		test.Assert(t, c.desc+": node content", c.expContent, node.Content)
	}
}

func TestNode_Seek(t *testing.T) {
	node := &Node{
		size:    3,
		Content: []byte("123"),
	}

	cases := []struct {
		expError error
		desc     string
		offset   int64
		whence   int
		exp      int64
	}{{
		desc:     "With invalid whence",
		offset:   5,
		whence:   3,
		expError: errWhence,
	}, {
		desc:     "With invalid offset",
		offset:   -5,
		whence:   io.SeekStart,
		expError: errOffset,
	}, {
		desc:   "SeekStart",
		offset: 5,
		whence: io.SeekStart,
		exp:    5,
	}, {
		desc:   "SeekCurrent",
		offset: 5,
		whence: io.SeekCurrent,
		exp:    10,
	}, {
		desc:   "SeekEnd",
		offset: 5,
		whence: io.SeekEnd,
		exp:    8,
	}}
	for _, c := range cases {
		t.Log(c.desc)

		got, err := node.Seek(c.offset, c.whence)

		test.Assert(t, "Seek", c.exp, got)
		test.Assert(t, "Seek error", c.expError, err)
	}
}
