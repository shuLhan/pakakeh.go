// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func ExampleDirWatcher() {
	var (
		ns      NodeState
		rootDir string
		err     error
	)

	rootDir, err = os.MkdirTemp("", "libmemfs")
	if err != nil {
		log.Fatal(err)
	}

	// In this example, we watch sub directory "assets" and its contents,
	// include only file with .adoc extension and ignoring files with
	// .html extension.
	dw := &DirWatcher{
		Options: Options{
			Root: rootDir,
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
		log.Fatal(err)
	}

	fmt.Printf("Deleting the root directory:\n")
	err = os.Remove(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s\n", ns.State, ns.Node.Path)

	// Create the root directory back with sub directory
	// This will trigger two FileStateCreated events, one for "/" and one
	// for "/assets".
	dirAssets := filepath.Join(rootDir, "assets")
	fmt.Printf("Re-create root directory with sub-directory:\n")
	err = os.MkdirAll(dirAssets, 0770)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s\n", ns.State, ns.Node.Path)
	ns = <-dw.C
	fmt.Printf("-- %s %s\n", ns.State, ns.Node.Path)

	// Modify the permission on root directory
	fmt.Printf("Chmod on root directory:\n")
	err = os.Chmod(rootDir, 0700)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	newFile := filepath.Join(rootDir, "new.adoc")
	fmt.Println("Create new file on root directory: /new.adoc")
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	fmt.Println("Remove file on root directory: /new.adoc")
	err = os.Remove(newFile)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	// Create sub-directory.
	subDir := filepath.Join(rootDir, "subdir")
	fmt.Println("Create new sub-directory: /subdir")
	err = os.Mkdir(subDir, 0770)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	// Add new file in sub directory.
	newFile = filepath.Join(subDir, "new.adoc")
	fmt.Println("Create new file in sub directory: /subdir/new.adoc")
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	fmt.Println("Remove file in sub directory: /subdir/new.adoc")
	err = os.Remove(newFile)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	// Creating file that is excluded should not trigger event.
	newFile = filepath.Join(subDir, "new.html")
	fmt.Println("Create excluded file in sub directory: /subdir/new.html")
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// Create file without extension in directory "assets" should trigger
	// event.
	newFile = filepath.Join(dirAssets, "new")
	fmt.Println("Create new file under assets: /assets/new")
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Printf("-- %s %s %s\n", ns.State, ns.Node.Path, ns.Node.Mode())

	dw.Stop()

	//Output:
	//Deleting the root directory:
	//-- FileStateDeleted /
	//Re-create root directory with sub-directory:
	//-- FileStateCreated /
	//-- FileStateCreated /assets
	//Chmod on root directory:
	//-- FileStateUpdateMode / drwx------
	//Create new file on root directory: /new.adoc
	//-- FileStateCreated /new.adoc -rw-------
	//Remove file on root directory: /new.adoc
	//-- FileStateDeleted /new.adoc -rw-------
	//Create new sub-directory: /subdir
	//-- FileStateCreated /subdir drwxr-x---
	//Create new file in sub directory: /subdir/new.adoc
	//-- FileStateCreated /subdir/new.adoc -rw-------
	//Remove file in sub directory: /subdir/new.adoc
	//-- FileStateDeleted /subdir/new.adoc -rw-------
	//Create excluded file in sub directory: /subdir/new.html
	//Create new file under assets: /assets/new
	//-- FileStateCreated /assets/new -rw-------
}
