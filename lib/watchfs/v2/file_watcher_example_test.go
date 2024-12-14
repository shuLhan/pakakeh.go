// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2"
)

func ExampleWatchFile() {
	var (
		name = `file.txt`
		opts = watchfs.FileWatcherOptions{
			File:     filepath.Join(os.TempDir(), name),
			Interval: 50 * time.Millisecond,
		}
	)

	fwatch := watchfs.WatchFile(opts)

	// On create ...
	_, err := os.Create(opts.File)
	if err != nil {
		log.Fatal(err)
	}

	var fi os.FileInfo = <-fwatch.C
	fmt.Printf("file %q created\n", fi.Name())

	// On update ...
	err = os.WriteFile(opts.File, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	fi = <-fwatch.C
	fmt.Printf("file %q updated\n", fi.Name())

	// On delete ...
	err = os.Remove(opts.File)
	if err != nil {
		log.Fatal(err)
	}

	fi = <-fwatch.C
	fmt.Printf("file deleted: %v\n", fi)

	fwatch.Stop()

	// Output:
	// file "file.txt" created
	// file "file.txt" updated
	// file deleted: <nil>
}
