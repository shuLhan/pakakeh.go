// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestPathFold(t *testing.T) {
	type testCase struct {
		path    string
		expPath string
	}

	var (
		userHomeDir string
		err         error
	)
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	var listTestCase = []testCase{{
		path:    filepath.Join(userHomeDir, `tmp`),
		expPath: `~/tmp`,
	}, {
		// Unclean path.
		path:    `//` + userHomeDir + `///tmp`,
		expPath: `~/tmp`,
	}}

	var (
		c       testCase
		gotPath string
	)
	for _, c = range listTestCase {
		gotPath, err = PathFold(c.path)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.path, c.expPath, gotPath)
	}
}

func TestPathUnfold(t *testing.T) {
	type testCase struct {
		path    string
		expPath string
	}

	var (
		username = os.Getenv(`USER`)

		workDir     string
		userHomeDir string
		err         error
	)

	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	workDir, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	var listTestCase = []testCase{{
		path:    `~/tmp`,
		expPath: filepath.Join(userHomeDir, `tmp`),
	}, {
		path:    `/home/user/~/tmp`,
		expPath: `/home/user/~/tmp`,
	}, {
		path:    `$HOME/tmp`,
		expPath: filepath.Join(userHomeDir, `tmp`),
	}, {
		path:    `/tmp/$USER/adir`,
		expPath: filepath.Join(`/`, `tmp`, username, `adir`),
	}, {
		path:    `~/$PWD/adir`,
		expPath: filepath.Join(userHomeDir, workDir, `adir`),
	}}

	var (
		c       testCase
		gotPath string
	)
	for _, c = range listTestCase {
		gotPath, err = PathUnfold(c.path)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.path, c.expPath, gotPath)
	}
}
