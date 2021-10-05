// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/memfs"
)

func TestDirWatcher(t *testing.T) {
	var wg sync.WaitGroup

	dir, err := ioutil.TempDir("", "libio")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	fmt.Printf(">>> Watching directory %q for changes ...\n", dir)

	exps := []struct {
		state FileState
		path  string
	}{{
		state: FileStateDeleted,
		path:  "/",
	}, {
		state: FileStateCreated,
		path:  "/",
	}, {
		state: FileStateUpdateMode,
		path:  "/",
	}, {
		state: FileStateCreated,
		path:  "/new.adoc",
	}, {
		state: FileStateDeleted,
		path:  filepath.Join(filepath.Base(dir), "/new.adoc"),
	}, {
		state: FileStateCreated,
		path:  "/sub",
	}, {
		state: FileStateCreated,
		path:  "/sub/new.adoc",
	}, {
		state: FileStateDeleted,
		path:  "sub/new.adoc",
	}, {
		state: FileStateCreated,
		path:  "/assets/new",
	}, {
		state: FileStateDeleted,
		path:  "assets/new",
	}}

	var x int32

	dw := &DirWatcher{
		Options: memfs.Options{
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
		Callback: func(ns *NodeState) {
			localx := atomic.LoadInt32(&x)
			if exps[localx].path != ns.Node.Path {
				log.Fatalf("TestDirWatcher got node path %q, want %q\n", ns.Node.Path, exps[x].path)
			}
			if exps[localx].state != ns.State {
				log.Fatalf("TestDirWatcher got state %d, want %d\n", ns.State, exps[x].state)
			}
			atomic.AddInt32(&x, 1)
			wg.Done()
		},
	}

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Delete the directory being watched.
	t.Logf("Deleting root directory %q ...\n", dir)
	wg.Add(1)
	err = os.Remove(dir)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Create the watched directory back with sub directory
	dirAssets := filepath.Join(dir, "assets")
	t.Logf("Re-create root directory %q ...\n", dirAssets)
	wg.Add(1)
	err = os.MkdirAll(dirAssets, 0770)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Modify the permission on root directory
	wg.Add(1)
	t.Logf("Modify root directory %q ...\n", dir)
	err = os.Chmod(dir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Add new file to watched directory.
	newFile := filepath.Join(dir, "new.adoc")
	t.Logf("Create new file on root directory: %q ...\n", newFile)
	wg.Add(1)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Remove file.
	t.Logf("Remove file on root directory: %q ...\n", newFile)
	wg.Add(1)
	err = os.Remove(newFile)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Create sub-directory.
	subDir := filepath.Join(dir, "sub")
	t.Logf("Create new sub-directory: %q ...\n", subDir)
	wg.Add(1)
	err = os.Mkdir(subDir, 0770)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Add new file in sub directory.
	newFile = filepath.Join(subDir, "new.adoc")
	t.Logf("Create new file in sub directory: %q ...\n", newFile)
	wg.Add(1)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Remove file in sub directory.
	t.Logf("Remove file in sub directory: %q ...\n", newFile)
	wg.Add(1)
	err = os.Remove(newFile)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

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
	wg.Add(1)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	wg.Add(1)
	dw.Stop()
}
