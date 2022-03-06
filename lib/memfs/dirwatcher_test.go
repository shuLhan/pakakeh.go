// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestDirWatcher(t *testing.T) {
	var (
		err   error
		gotNS NodeState
		dir   string
		x     int
	)

	dir = t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf(">>> Watching directory %q for changes ...\n", dir)

	exps := []struct {
		path  string
		state FileState
	}{{
		state: FileStateDeleted,
		path:  "/",
	}, {
		state: FileStateCreated,
		path:  "/",
	}, {
		state: FileStateCreated,
		path:  "/assets",
	}, {
		state: FileStateUpdateMode,
		path:  "/",
	}, {
		state: FileStateCreated,
		path:  "/new.adoc",
	}, {
		state: FileStateDeleted,
		path:  "/new.adoc",
	}, {
		state: FileStateCreated,
		path:  "/sub",
	}, {
		state: FileStateCreated,
		path:  "/sub/new.adoc",
	}, {
		state: FileStateDeleted,
		path:  "/sub/new.adoc",
	}, {
		state: FileStateCreated,
		path:  "/assets/new",
	}, {
		state: FileStateDeleted,
		path:  "/assets/new",
	}}

	dw := &DirWatcher{
		Options: Options{
			Root: dir,
			Includes: []string{
				`assets/.*`,
				`.*\.adoc$`,
			},
			Excludes: []string{
				`.*\.html$`,
			},
		},
		Delay: 150 * time.Millisecond,
	}

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Delete the directory being watched.
	t.Logf("Deleting root directory %q ...\n", dir)
	err = os.Remove(dir)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Create the watched directory back with sub directory
	// This will trigger two FileStateCreated events, one for "/" and one
	// for "/assets".
	dirAssets := filepath.Join(dir, "assets")
	t.Logf("Re-create root directory %q ...\n", dirAssets)
	err = os.MkdirAll(dirAssets, 0770)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Modify the permission on root directory
	t.Logf("Modify root directory %q ...\n", dir)
	err = os.Chmod(dir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Add new file to watched directory.
	newFile := filepath.Join(dir, "new.adoc")
	t.Logf("Create new file on root directory: %q ...\n", newFile)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Remove file.
	t.Logf("Remove file on root directory: %q ...\n", newFile)
	err = os.Remove(newFile)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Create sub-directory.
	subDir := filepath.Join(dir, "sub")
	t.Logf("Create new sub-directory: %q ...\n", subDir)
	err = os.Mkdir(subDir, 0770)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Add new file in sub directory.
	newFile = filepath.Join(subDir, "new.adoc")
	t.Logf("Create new file in sub directory: %q ...\n", newFile)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Remove file in sub directory.
	t.Logf("Remove file in sub directory: %q ...\n", newFile)
	err = os.Remove(newFile)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	// Create exclude file, should not trigger event.
	newFile = filepath.Join(subDir, "new.html")
	t.Logf("Create excluded file in sub directory: %q ...\n", newFile)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Create file without extension in white list directory "assets",
	// should trigger event.
	newFile = filepath.Join(dirAssets, "new")
	t.Logf("Create new file on assets: %q ...\n", newFile)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	gotNS = <-dw.C
	test.Assert(t, "path", exps[x].path, gotNS.Node.Path)
	test.Assert(t, "state", exps[x].state, gotNS.State)
	x++

	dw.Stop()
}

//
// Test renaming sub-directory being watched.
//
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

	var expDirs = []string{
		"/newsubdir",
	}

	test.Assert(t, "dirs", expDirs, dw.dirsKeys())
}
