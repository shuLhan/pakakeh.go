// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>

package os

import (
	"bytes"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestConfirmYesNo(t *testing.T) {
	type testCase struct {
		answer   string
		defIsYes bool
		exp      bool
	}

	var cases = []testCase{{
		defIsYes: true,
		exp:      true,
	}, {
		defIsYes: true,
		answer:   `  `,
		exp:      true,
	}, {
		defIsYes: true,
		answer:   `  no`,
		exp:      false,
	}, {
		defIsYes: true,
		answer:   ` yes`,
		exp:      true,
	}, {
		defIsYes: true,
		answer:   ` Ys`,
		exp:      true,
	}, {
		defIsYes: false,
		exp:      false,
	}, {
		defIsYes: false,
		answer:   ``,
		exp:      false,
	}, {

		defIsYes: false,
		answer:   `  no`,
		exp:      false,
	}, {
		defIsYes: false,
		answer:   `  yes`,
		exp:      true,
	}}

	var (
		mockReader bytes.Buffer
		c          testCase
		got        bool
	)

	for _, c = range cases {
		t.Log(c)
		mockReader.Reset()

		// Write the answer to be read.
		mockReader.WriteString(c.answer + "\n")

		got = ConfirmYesNo(&mockReader, `confirm`, c.defIsYes)

		test.Assert(t, `answer`, c.exp, got)
	}
}

func TestCopy(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		out    string
		expErr string
	}{{
		desc:   `Without output file`,
		in:     `testdata/Copy/input.txt`,
		expErr: `Copy: failed to open output file: open : no such file or directory`,
	}, {
		desc:   `Without input file`,
		out:    `testdata/Copy/output.txt`,
		expErr: `Copy: failed to open input file: open : no such file or directory`,
	}, {
		desc: `With input and output`,
		in:   `testdata/Copy/input.txt`,
		out:  `testdata/Copy/output.txt`,
	}}

	for _, c := range cases {
		err := Copy(c.out, c.in)
		if err != nil {
			test.Assert(t, c.desc, c.expErr, err.Error())
			continue
		}

		exp, err := os.ReadFile(c.in)
		if err != nil {
			t.Fatal(err)
		}

		got, err := os.ReadFile(c.out)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc, string(exp), string(got))
	}
}

func TestIsBinaryStream(t *testing.T) {
	listCase := []struct {
		path string
		exp  bool
	}{{
		path: `testdata/exp.bz2`,
		exp:  true,
	}, {
		path: `os.go`,
		exp:  false,
	}}
	for _, tc := range listCase {
		content, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		got := IsBinaryStream(content)
		test.Assert(t, tc.path, tc.exp, got)
	}
}

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

		test.Assert(t, "", c.exp, got)
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
		relpath: "testdata/.empty",
		exp:     true,
	}, {
		desc:    "With file exist",
		parent:  wd,
		relpath: "testdata/.empty",
		exp:     true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsFileExist(c.parent, c.relpath)

		test.Assert(t, "", c.exp, got)
	}
}

func TestRmdirEmptyAll(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Remove(`testdata/file`)
		_ = os.RemoveAll(`testdata/a`)
		_ = os.RemoveAll(`testdata/dirempty`)
	})

	cases := []struct {
		desc        string
		createDir   string
		createFile  string
		path        string
		expExist    string
		expNotExist string
	}{{
		desc:       `With path as file`,
		path:       `testdata/file`,
		createFile: `testdata/file`,
		expExist:   `testdata/file`,
	}, {
		desc:      `With empty path`,
		createDir: `testdata/a/b/c/d`,
		expExist:  `testdata/a/b/c/d`,
	}, {
		desc:        `With non empty at middle`,
		createDir:   `testdata/a/b/c/d`,
		createFile:  `testdata/a/b/file`,
		path:        `testdata/a/b/c/d`,
		expExist:    `testdata/a/b/file`,
		expNotExist: `testdata/a/b/c`,
	}, {
		desc:        `With first path not exist`,
		createDir:   `testdata/a/b/c`,
		path:        `testdata/a/b/c/d`,
		expExist:    `testdata/a/b/file`,
		expNotExist: `testdata/a/b/c`,
	}, {
		desc:        `With non empty at parent`,
		createDir:   `testdata/dirempty/a/b/c/d`,
		path:        `testdata/dirempty/a/b/c/d`,
		expExist:    `testdata`,
		expNotExist: `testdata/dirempty`,
	}}

	var (
		err error
		f   *os.File
	)
	for _, c := range cases {
		t.Log(c.desc)

		if len(c.createDir) > 0 {
			err = os.MkdirAll(c.createDir, 0700)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.createFile) > 0 {
			f, err = os.Create(c.createFile)
			if err != nil {
				t.Fatal(err)
			}
			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}

		err = RmdirEmptyAll(c.path)
		if err != nil {
			t.Fatal(err)
		}

		if len(c.expExist) > 0 {
			_, err = os.Stat(c.expExist)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.expNotExist) > 0 {
			_, err = os.Stat(c.expNotExist)
			if !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	}
}

// TestStat test to see the difference between Stat and Lstat.
func TestStat(t *testing.T) {
	t.Skip()

	var (
		files = []string{
			`testdata/exp`,
			`testdata/symlink_file`,
			`testdata/symlink_symlink_file`,
			`testdata/exp_dir`,
			`testdata/symlink_dir`,
			`testdata/symlink_symlink_dir`,
		}

		fi   os.FileInfo
		file string
		err  error
	)

	for _, file = range files {
		fi, err = os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf(` Stat %s: mode:%s, size:%d, is_dir:%t, is_symlink:%t, modtime:%s`,
			file, fi.Mode(), fi.Size(), fi.IsDir(),
			fi.Mode()&os.ModeSymlink != 0, fi.ModTime())

		fi, err = os.Lstat(file)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf(`Lstat %s: mode:%s, size:%d, is_dir:%t, is_symlink:%t, modtime:%s`,
			file, fi.Mode(), fi.Size(), fi.IsDir(),
			fi.Mode()&os.ModeSymlink != 0, fi.ModTime())
	}
}
