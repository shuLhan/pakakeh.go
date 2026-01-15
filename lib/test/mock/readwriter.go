// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package mock

import (
	"bytes"
)

// ReadWriter mock for testing with io.ReadWriter or io.StringWriter.
// Every call to Read will read from BufRead and every call to Write or
// WriteString will write to BufWrite.
type ReadWriter struct {
	BufRead  bytes.Buffer
	BufWrite bytes.Buffer
}

// Read read a stream of byte from read buffer.
func (rw *ReadWriter) Read(b []byte) (n int, err error) {
	return rw.BufRead.Read(b)
}

// Write write a stream of byte into write buffer.
func (rw *ReadWriter) Write(b []byte) (n int, err error) {
	return rw.BufWrite.Write(b)
}

// WriteString write a string into write buffer.
func (rw *ReadWriter) WriteString(s string) (n int, err error) {
	return rw.BufWrite.WriteString(s)
}

// Reset reset read and write buffers.
func (rw *ReadWriter) Reset() {
	rw.BufRead.Reset()
	rw.BufWrite.Reset()
}
