// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

//
// Node represent the data returned from Readlink.
//
type Node struct {
	FileName string
	LongName string
	Attrs    *FileAttrs
}
