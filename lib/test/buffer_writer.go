// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package test

import (
	"bytes"
	"fmt"
)

// BufferWriter implement the Writer interface.
// Any call to ErrorXxx, FatalXxx, and LogXxx will write to embedded
// bytes.Buffer.
type BufferWriter struct {
	bytes.Buffer
}

// Error write the arguments into buffer.
func (bw *BufferWriter) Error(args ...any) {
	fmt.Fprintln(bw, args...)
}

// Errorf write formatted string with arguments into buffer.
func (bw *BufferWriter) Errorf(format string, args ...any) {
	fmt.Fprintf(bw, format, args...)
}

// Fatal write the arguments to buffer.
func (bw *BufferWriter) Fatal(args ...any) {
	fmt.Fprint(bw, args...)
}

// Fatalf write formatted string with arguments into buffer.
func (bw *BufferWriter) Fatalf(format string, args ...any) {
	fmt.Fprintf(bw, format, args...)
}

func (bw *BufferWriter) Helper() {
	// NOOP
}

// Log write the arguments into buffer.
func (bw *BufferWriter) Log(args ...any) {
	fmt.Fprint(bw, args...)
}

// Logf write formatted string with arguments into buffer.
func (bw *BufferWriter) Logf(format string, args ...any) {
	fmt.Fprintf(bw, format, args...)
}
