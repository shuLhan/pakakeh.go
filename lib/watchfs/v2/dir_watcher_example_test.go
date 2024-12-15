// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2"
)

func ExampleWatchDir() {
	var (
		dirTemp string
		err     error
	)
	dirTemp, err = os.MkdirTemp(``, ``)
	if err != nil {
		log.Fatal(err)
	}

	var (
		fileToWatch = filepath.Join(dirTemp, `.rescan`)
		opts        = watchfs.DirWatcherOptions{
			FileWatcherOptions: watchfs.FileWatcherOptions{
				File:     fileToWatch,
				Interval: 50 * time.Millisecond,
			},
			Root:     dirTemp,
			Includes: []string{`.*\.adoc$`},
			Excludes: []string{`exc$`, `.*\.html$`},
		}
		dwatch *watchfs.DirWatcher
	)

	dwatch, err = watchfs.WatchDir(opts)
	if err != nil {
		log.Fatal(err)
	}

	var (
		fileAadoc = filepath.Join(opts.Root, `a.adoc`)
		fileBadoc = filepath.Join(opts.Root, `b.adoc`)
		fileAhtml = filepath.Join(opts.Root, `a.html`)
	)

	err = os.WriteFile(fileAadoc, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(fileAhtml, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(fileBadoc, nil, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// Write to the file that we watch for changes to trigger rescan.
	err = os.WriteFile(fileToWatch, []byte(`x`), 0600)
	if err != nil {
		log.Fatal(err)
	}

	var changes []os.FileInfo = <-dwatch.C
	var names []string
	for _, fi := range changes {
		// Since we use temporary directory, print only the base
		// name to make it works on all system.
		names = append(names, filepath.Base(fi.Name()))
	}
	sort.Strings(names)
	fmt.Println(names)
	// Output:
	// [a.adoc b.adoc]
}
