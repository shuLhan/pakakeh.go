// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"os"
	"path/filepath"
)

// RmdirEmptyAll remove directory in path if it's empty until one of the
// parent is not empty.
//
// DEPRECATED: moved to [lib/os#RmdirEmptyAll].
func RmdirEmptyAll(path string) error {
	if len(path) == 0 {
		return nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return RmdirEmptyAll(filepath.Dir(path))
	}
	if !fi.IsDir() {
		return nil
	}
	if !IsDirEmpty(path) {
		return nil
	}
	err = os.Remove(path)
	if err != nil {
		return err
	}

	return RmdirEmptyAll(filepath.Dir(path))
}
