// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDirWatcher(t *testing.T) {
	var wg sync.WaitGroup

	dir, err := ioutil.TempDir("", "libio")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)
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
		state: FileStateModified,
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
	}}

	expIdx := 0

	dw := &DirWatcher{
		Path:  dir,
		Delay: 150 * time.Millisecond,
		Includes: []string{
			`assets/.*`,
			`.*\.adoc$`,
		},
		Excludes: []string{
			`.*\.html$`,
		},
		Callback: func(ns *NodeState) {
			tt := t
			if exps[expIdx].path != ns.Node.Path {
				tt.Fatalf("Callback got node path %q, want %q\n", ns.Node.Path, exps[expIdx].path)
			}
			if exps[expIdx].state != ns.State {
				tt.Fatalf("Callback got state %d, want %d\n", ns.State, exps[expIdx].state)
			}
			expIdx++
			wg.Done()
		},
	}

	err = dw.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Delete the directory being watched.
	fmt.Printf(">>> Deleting root directory %q ...\n", dir)
	wg.Add(1)
	err = os.Remove(dir)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Create the watched directory back with sub directory
	dirAssets := filepath.Join(dir, "assets")
	fmt.Printf(">>> Re-create root directory %q ...\n", dirAssets)
	wg.Add(1)
	err = os.MkdirAll(dirAssets, 0770)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Modify the permission on root directory
	wg.Add(1)
	fmt.Printf(">>> Modify root directory %q ...\n", dir)
	err = os.Chmod(dir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Add new file to watched directory.
	newFile := filepath.Join(dir, "new.adoc")
	fmt.Printf(">>> Create new file %q ...\n", newFile)
	wg.Add(1)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Remove file.
	fmt.Printf(">>> Remove file %q ...\n", newFile)
	wg.Add(1)
	err = os.Remove(newFile)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Create sub-directory.
	subDir := filepath.Join(dir, "sub")
	fmt.Printf(">>> Create new sub-directory %q ...\n", subDir)
	wg.Add(1)
	err = os.Mkdir(subDir, 0770)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Add new file in sub directory.
	newFile = filepath.Join(subDir, "new.adoc")
	fmt.Printf(">>> Create new file %q ...\n", newFile)
	wg.Add(1)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Remove file in sub directory.
	fmt.Printf(">>> Remove file %q ...\n", newFile)
	wg.Add(1)
	err = os.Remove(newFile)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	// Create exclude file, should not trigger event.
	newFile = filepath.Join(subDir, "new.html")
	fmt.Printf(">>> Create exclude file %q ...\n", newFile)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Create file without extension in sub directory, should not trigger
	// event.
	newFile = filepath.Join(subDir, "new")
	fmt.Printf(">>> Create new file %q ...\n", newFile)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Create file without extension in white list directory "assets",
	// should trigger event.
	newFile = filepath.Join(dirAssets, "new")
	fmt.Printf(">>> Create new file %q ...\n", newFile)
	wg.Add(1)
	err = ioutil.WriteFile(newFile, nil, 0600)
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	dw.Stop()
}
