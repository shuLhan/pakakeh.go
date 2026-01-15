// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import (
	"fmt"
	"os"
)

// StdoutForwarder write the metrics to os.Stdout.
// This type is used as example and to provide wrapper for os.Stdout, since
// user should not call Close on os.Stdout.
type StdoutForwarder struct {
	formatter Formatter
}

// NewStdoutForwarder create new StdoutForwarder using f as Formatter.
func NewStdoutForwarder(f Formatter) *StdoutForwarder {
	return &StdoutForwarder{
		formatter: f,
	}
}

// Close on StdoutForwarder sync the Stdout.
func (stdout *StdoutForwarder) Close() error {
	os.Stdout.Sync()
	return nil
}

// Formatter return the Formatter used by this StdoutForwarder.
func (stdout *StdoutForwarder) Formatter() Formatter {
	return stdout.formatter
}

// Write the raw metrics to stdout.
func (stdout *StdoutForwarder) Write(wire []byte) (n int, err error) {
	n, err = os.Stdout.Write(wire)
	if err != nil {
		return n, fmt.Errorf(`StdoutForwarder.Forward: %w`, err)
	}
	return n, nil
}
