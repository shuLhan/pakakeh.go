// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

// Test renaming sub-directory being watched.
func TestDirWatcher_renameDirectory(t *testing.T) {
	var (
		dw  DirWatcher
		err error

		rootDir    string
		subDir     string
		subDirFile string
		newSubDir  string
	)

	//
	// Create a directory with its content to be watched.
	//
	//	rootDir
	//	|_ subDir
	//	   |_ subDirFile
	//

	rootDir = t.TempDir()

	subDir = filepath.Join(rootDir, "subdir")
	err = os.Mkdir(subDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	subDirFile = filepath.Join(subDir, "testfile")
	err = os.WriteFile(subDirFile, []byte(`content of testfile`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	dw = DirWatcher{
		Options: Options{
			Root: rootDir,
		},
		Delay: 200 * time.Millisecond,
	}

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Wait for all watcher started.
	time.Sleep(400 * time.Millisecond)

	newSubDir = filepath.Join(rootDir, "newsubdir")
	err = os.Rename(subDir, newSubDir)
	if err != nil {
		t.Fatal(err)
	}

	<-dw.C
	<-dw.C
	<-dw.C

	dw.Stop()

	var expDirs = []string{
		"/newsubdir",
	}

	test.Assert(t, "dirs", expDirs, dw.dirsKeys())
}

func TestDirWatcher_removeDirSymlink(t *testing.T) {
	var (
		dirWd string
		err   error
	)

	dirWd, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	var (
		dirTmp  = t.TempDir()
		dirSub  = filepath.Join(dirTmp, `sub`)
		fileOld = filepath.Join(dirWd, `testdata`, `index.html`)
		fileNew = filepath.Join(dirSub, `index.html`)
		opts    = Options{
			Root: dirTmp,
		}
		dw = DirWatcher{
			Options: opts,
			Delay:   200 * time.Millisecond,
		}

		got NodeState
	)

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Wait for all watcher started.
	time.Sleep(500 * time.Millisecond)

	err = os.Mkdir(dirSub, 0700)
	if err != nil {
		t.Fatal(err)
	}
	got = <-dw.C
	test.Assert(t, `Mkdir state`, FileStateCreated, got.State)
	test.Assert(t, `Mkdir path`, `/sub`, got.Node.Path)

	err = os.Symlink(fileOld, fileNew)
	if err != nil {
		t.Fatal(err)
	}
	got = <-dw.C
	test.Assert(t, `Symlink state`, FileStateCreated, got.State)
	test.Assert(t, `Symlink path`, `/sub/index.html`, got.Node.Path)

	err = os.RemoveAll(dirSub)
	if err != nil {
		t.Fatal(err)
	}
	got = <-dw.C
	test.Assert(t, `RemoveAll state`, FileStateDeleted, got.State)
	test.Assert(t, `RemoveAll path`, `/sub/index.html`, got.Node.Path)
}
