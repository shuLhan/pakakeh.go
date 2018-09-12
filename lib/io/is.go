// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io"
	"os"
	"path/filepath"
)

//
// IsDirEmpty will return true if directory is not exist or empty; otherwise
// it will return false.
//
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

//
// IsFileExist will return true if relative path is exist on parent directory;
// otherwise it will return false.
//
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
