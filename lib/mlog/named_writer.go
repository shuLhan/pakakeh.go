// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
