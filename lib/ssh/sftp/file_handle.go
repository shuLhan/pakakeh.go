// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

type FileHandle struct {
	remotePath string // The remote path.
	v          []byte // The handle value returned from open().
}
