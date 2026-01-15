// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import (
	"bytes"
	"fmt"
	"sync"
)

// BufferForwarder write the metrics to underlying [bytes.Buffer].
type BufferForwarder struct {
	formatter Formatter
	bb        bytes.Buffer
	sync.Mutex
}

// NewBufferForwarder create new BufferForwarder using f as Formatter.
func NewBufferForwarder(f Formatter) *BufferForwarder {
	return &BufferForwarder{
		formatter: f,
	}
}

// Bytes return the metrics that has been written to Buffer.
// Once this method called the underlying Buffer will be resetted.
func (buf *BufferForwarder) Bytes() (b []byte) {
	buf.Lock()
	b = buf.bb.Bytes()
	buf.bb.Reset()
	buf.Unlock()
	return b
}

// Close on Buffer is a no-op.
func (buf *BufferForwarder) Close() error {
	return nil
}

// Formatter return the Formatter used by this BufferForwarder.
func (buf *BufferForwarder) Formatter() Formatter {
	return buf.formatter
}

// Write the raw metrics to Buffer.
func (buf *BufferForwarder) Write(wire []byte) (n int, err error) {
	buf.Lock()
	defer buf.Unlock()

	n, err = buf.bb.Write(wire)
	if err != nil {
		return n, fmt.Errorf(`BufferForwarder.Forward: %w`, err)
	}
	return n, nil
}
