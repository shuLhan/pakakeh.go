// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs"
)

func ExampleNewWatcher() {
	var (
		content = `Content of file`

		f       *os.File
		watcher *watchfs.Watcher
		ns      watchfs.NodeState
		err     error
	)

	// Create a file to be watched.
	f, err = os.CreateTemp(``, `watcher`)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err = watchfs.NewWatcher(f.Name(), 150*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	// Update file mode.
	err = f.Chmod(0700)
	if err != nil {
		log.Fatal(err)
	}

	ns = <-watcher.C
	fmt.Printf("State: %s\n", ns.State)
	fmt.Printf("File mode: %s\n", ns.Node.Mode())
	fmt.Printf("File size: %d\n", ns.Node.Size())

	// Update content of file.
	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	ns = <-watcher.C
	fmt.Printf("State: %s\n", ns.State)
	fmt.Printf("File mode: %s\n", ns.Node.Mode())
	fmt.Printf("File size: %d\n", ns.Node.Size())

	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Remove the file.
	err = os.Remove(f.Name())
	if err != nil {
		log.Fatal(err)
	}
	ns = <-watcher.C
	fmt.Printf("State: %s\n", ns.State)
	fmt.Printf("File mode: %s\n", ns.Node.Mode())
	fmt.Printf("File size: %d\n", ns.Node.Size())

	// Output:
	// State: FileStateUpdateMode
	// File mode: -rwx------
	// File size: 0
	// State: FileStateUpdateContent
	// File mode: -rwx------
	// File size: 15
	// State: FileStateDeleted
	// File mode: -rwx------
	// File size: 15
}
