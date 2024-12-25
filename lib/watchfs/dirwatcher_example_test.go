// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs"
)

func ExampleDirWatcher() {
	var (
		rootDir string
		err     error
	)

	rootDir, err = os.MkdirTemp(``, `ExampleDirWatcher`)
	if err != nil {
		log.Fatal(err)
	}

	// In this example, we watch sub directory "assets" and its
	// contents, including only files with ".adoc" extension and
	// excluding files with ".html" extension.
	var dw = &watchfs.DirWatcher{
		Options: watchfs.DirWatcherOptions{
			Root: rootDir,
			Includes: []string{
				`assets/.*`,
				`.*\.adoc$`,
			},
			Excludes: []string{
				`.*\.html$`,
			},
			Delay: 100 * time.Millisecond,
		},
	}

	err = dw.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(`Deleting the root directory:`)
	err = os.Remove(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	var ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path)

	// Create the root directory back with sub directory
	// This will trigger one FileStateCreated event, for "/".
	fmt.Println(`Re-create root directory with sub-directory:`)
	var dirAssets = filepath.Join(rootDir, `assets`)
	err = os.MkdirAll(dirAssets, 0770)
	if err != nil {
		log.Fatal(err)
	}

	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path)

	// Modify the permission on root directory
	fmt.Println(`Chmod on root directory:`)
	err = os.Chmod(rootDir, 0700)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	fmt.Println(`Create new file on root directory: /new.adoc`)
	var newFile = filepath.Join(rootDir, `new.adoc`)
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	fmt.Println(`Remove file on root directory: /new.adoc`)
	err = os.Remove(newFile)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	fmt.Println(`Create new sub-directory: /subdir`)
	var subDir = filepath.Join(rootDir, `subdir`)
	err = os.Mkdir(subDir, 0770)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	// Add new file in sub directory.
	newFile = filepath.Join(subDir, `new.adoc`)
	fmt.Println(`Create new file in sub directory: /subdir/new.adoc`)
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	fmt.Println(`Remove file in sub directory: /subdir/new.adoc`)
	err = os.Remove(newFile)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	// Creating file that is excluded should not trigger event.
	fmt.Println(`Create excluded file in sub directory: /subdir/new.html`)
	newFile = filepath.Join(subDir, `new.html`)
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// Create file without extension in directory "assets" should trigger
	// event.
	newFile = filepath.Join(dirAssets, `new`)
	fmt.Println(`Create new file under assets: /assets/new`)
	err = os.WriteFile(newFile, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-dw.C
	fmt.Println(`--`, ns.State, ns.Node.Path, ns.Node.Mode())

	// TODO: fix data race.
	//dw.Stop()

	// Output:
	// Deleting the root directory:
	// -- FileStateDeleted /
	// Re-create root directory with sub-directory:
	// -- FileStateCreated /
	// Chmod on root directory:
	// -- FileStateUpdateMode / drwx------
	// Create new file on root directory: /new.adoc
	// -- FileStateCreated /new.adoc -rw-------
	// Remove file on root directory: /new.adoc
	// -- FileStateDeleted /new.adoc -rw-------
	// Create new sub-directory: /subdir
	// -- FileStateCreated /subdir drwxr-x---
	// Create new file in sub directory: /subdir/new.adoc
	// -- FileStateCreated /subdir/new.adoc -rw-------
	// Remove file in sub directory: /subdir/new.adoc
	// -- FileStateDeleted /subdir/new.adoc -rw-------
	// Create excluded file in sub directory: /subdir/new.html
	// Create new file under assets: /assets/new
	// -- FileStateCreated /assets/new -rw-------
}
