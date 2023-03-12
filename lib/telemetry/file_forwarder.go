// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

import (
	"fmt"
	"io"
)

// FileForwarder forward the raw metrics into file.
type FileForwarder struct {
	fmt  Formatter
	file io.WriteCloser
}

// NewFileForwarder create new FileForwarder using fmt as the Formatter.
func NewFileForwarder(fmt Formatter, file io.WriteCloser) (fwd *FileForwarder) {
	fwd = &FileForwarder{
		fmt:  fmt,
		file: file,
	}
	return fwd
}

// Close the underlying file.
// Calling Forward after closing Forwarder may cause panic.
func (fwd *FileForwarder) Close() (err error) {
	if fwd.file != nil {
		err = fwd.file.Close()
	}
	return err
}

// Formatter return the Formatter that is used by this FileForwarder.
func (fwd *FileForwarder) Formatter() Formatter {
	return fwd.fmt
}

// Forward write the formatted metrics into file.
func (fwd *FileForwarder) Forward(wire []byte) (err error) {
	_, err = fwd.file.Write(wire)
	if err != nil {
		return fmt.Errorf(`FileForwarder.Forward: %w`, err)
	}
	return nil
}
