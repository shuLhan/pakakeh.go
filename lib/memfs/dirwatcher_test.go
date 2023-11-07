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
		Delay: 100 * time.Millisecond,
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
			Delay:   100 * time.Millisecond,
		}

		got NodeState
	)

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Add delay for modtime to changes.
	time.Sleep(100 * time.Millisecond)

	err = os.Mkdir(dirSub, 0700)
	if err != nil {
		t.Fatal(err)
	}
	got = <-dw.C
	test.Assert(t, `Mkdir state`, FileStateCreated, got.State)
	test.Assert(t, `Mkdir path`, `/sub`, got.Node.Path)

	// Add delay for modtime to changes.
	time.Sleep(100 * time.Millisecond)

	err = os.Symlink(fileOld, fileNew)
	if err != nil {
		t.Fatal(err)
	}
	got = <-dw.C
	test.Assert(t, `Symlink state`, FileStateCreated, got.State)
	test.Assert(t, `Symlink path`, `/sub/index.html`, got.Node.Path)

	// Add delay for modtime to changes.
	time.Sleep(100 * time.Millisecond)

	err = os.RemoveAll(dirSub)
	if err != nil {
		t.Fatal(err)
	}
	got = <-dw.C
	test.Assert(t, `RemoveAll state`, FileStateDeleted, got.State)
	test.Assert(t, `RemoveAll path`, `/sub/index.html`, got.Node.Path)
}

func TestDirWatcher_withSymlink(t *testing.T) {
	// Initialize the file and directory for symlink.

	var (
		dirSource     = t.TempDir()
		dirDest       = t.TempDir()
		symlinkSource = filepath.Join(dirSource, `symlinkSource`)
		symlinkDest   = filepath.Join(dirDest, `symlinkDest`)
		data          = []byte(`content of symlink`)

		err error
	)

	err = os.WriteFile(symlinkSource, data, 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink(symlinkSource, symlinkDest)
	if err != nil {
		t.Fatal(err)
	}

	// Create the DirWatcher instance and start watching the changes.

	var dw = DirWatcher{
		Options: Options{
			Root: dirDest,
		},
		Delay: 100 * time.Millisecond,
	}

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Add delay for modtime to changes.
	time.Sleep(100 * time.Millisecond)

	var gotns NodeState

	// Write to file source.
	data = []byte(`new content of symlink`)
	err = os.WriteFile(symlinkSource, data, 0600)
	if err != nil {
		t.Fatal(err)
	}

	gotns = <-dw.C
	test.Assert(t, `path`, `/symlinkDest`, gotns.Node.Path)
	test.Assert(t, `state`, FileStateUpdateContent, gotns.State)

	// Add delay for modtime to changes.
	time.Sleep(100 * time.Millisecond)

	// Write to symlink file destination.
	data = []byte(`new content of symlink destination`)
	err = os.WriteFile(symlinkDest, data, 0600)
	if err != nil {
		t.Fatal(err)
	}

	gotns = <-dw.C
	test.Assert(t, `path`, `/symlinkDest`, gotns.Node.Path)
	test.Assert(t, `state`, FileStateUpdateContent, gotns.State)

	dw.Stop()
}
