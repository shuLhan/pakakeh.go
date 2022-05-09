// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

import "os"

// dirEntry represent the internal data returned from Readdir, Readlink, or
// Realpath.
type dirEntry struct {
	fileName string
	longName string
	attrs    *FileAttrs
}

func (de *dirEntry) Name() string {
	return de.fileName
}

func (de *dirEntry) IsDir() bool {
	return de.attrs.IsDir()
}

func (de *dirEntry) Type() os.FileMode {
	return os.FileMode(de.attrs.permissions & fileTypeMask)
}

func (de *dirEntry) Info() (os.FileInfo, error) {
	return de.attrs, nil
}
