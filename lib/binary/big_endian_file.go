// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"fmt"
	"os"
	"sync"
)

// BigEndianFile support reading and writing Go types into file with
// big-endian byte order.
// Unlike [binary.Write] in the standard library, the Write method in here
// support writing struct field with type slice and string.
type BigEndianFile struct {
	name string
	file *os.File

	// val is the backing storage for the file.
	val []byte

	mtx sync.Mutex
}

// OpenBigEndianFile open the file for read and write.
// It will return an error if file does not exists.
func OpenBigEndianFile(name string) (bef *BigEndianFile, err error) {
	var logp = `OpenBigEndianFile`

	bef = &BigEndianFile{
		name: name,
	}

	bef.file, err = os.OpenFile(name, os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var fi os.FileInfo
	fi, err = bef.file.Stat()
	if err != nil {
		goto fail
	}

	bef.val = make([]byte, fi.Size())

	// Read all contents.
	_, err = bef.file.Read(bef.val)
	if err != nil {
		goto fail
	}

	return bef, nil
fail:
	_ = bef.file.Close()
	return nil, fmt.Errorf(`%s: %w`, logp, err)
}

func (bef *BigEndianFile) Close() (err error) {
	bef.mtx.Lock()
	err = bef.file.Close()
	bef.mtx.Unlock()
	return err
}

func (bef *BigEndianFile) Read(data any) (err error) {
	return nil
}

func (bef *BigEndianFile) Seek(off int64, whence int) (n int64, err error) {
	return n, nil
}

func (bef *BigEndianFile) Write(data any) (n int64, err error) {
	return n, nil
}
