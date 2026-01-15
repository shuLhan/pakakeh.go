// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package sftp

// FileHandle define the container to store remote file.
type FileHandle struct {
	remotePath string // The remote path.
	v          []byte // The handle value returned from open().
}
