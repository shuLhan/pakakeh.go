// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io"
	"os"
	"path/filepath"

	"github.com/shuLhan/share/lib/ascii"
)

// IsBinary will return true if content of file is binary.
// If file is not exist or there is an error when reading or closing the file,
// it will return false.
//
// DEPRECATED: moved to [lib/os#IsBinary].
func IsBinary(file string) bool {
	var (
		f         *os.File
		err       error
		total     int
		printable int
	)

	f, err = os.Open(file)
	if err != nil {
		return false
	}

	var (
		content = make([]byte, 768)

		n int
		x int
	)

	for total < 512 {
		n, err = f.Read(content)
		if err != nil {
			break
		}

		content = content[:n]

		for x = 0; x < len(content); x++ {
			if ascii.IsSpace(content[x]) {
				continue
			}
			if content[x] >= 33 && content[x] <= 126 {
				printable++
			}
			total++
		}
	}

	err = f.Close()
	if err != nil {
		return false
	}

	var ratio = float64(printable) / float64(total)

	return ratio <= float64(0.75)
}

// IsDirEmpty will return true if directory is not exist or empty; otherwise
// it will return false.
//
// DEPRECATED: moved to [lib/os#IsDirEmpty].
func IsDirEmpty(dir string) (ok bool) {
	d, err := os.Open(dir)
	if err != nil {
		ok = true
		return
	}

	_, err = d.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			ok = true
		}
	}

	_ = d.Close()

	return
}

// IsFileExist will return true if relative path is exist on parent directory;
// otherwise it will return false.
//
// DEPRECATED: moved to [lib/os#IsFileExist].
func IsFileExist(parent, relpath string) bool {
	path := filepath.Join(parent, relpath)

	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}
