// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package mlog

import "io"

// NamedWriter is io.Writer with name.
type NamedWriter interface {
	Name() string
	Write(b []byte) (n int, err error)
}

type namedWriter struct {
	io.Writer
	name string
}

// Name return the log name.
func (nw *namedWriter) Name() string {
	return nw.name
}

// NewNamedWriter create new NamedWriter instance.
func NewNamedWriter(name string, w io.Writer) NamedWriter {
	return &namedWriter{
		Writer: w,
		name:   name,
	}
}
