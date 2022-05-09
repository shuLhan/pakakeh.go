// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dsv is a library for working with delimited separated value (DSV).
//
// DSV is a free-style form of Comma Separated Value (CSV) format of text data,
// where each row is separated by newline, and each column can be separated by
// any string enclosed with left-quote and right-quote.
package dsv

import (
	"errors"
)

const (
	// DefaultRejected define the default file which will contain the
	// rejected row.
	DefaultRejected = "rejected.dat"

	// DefaultMaxRows define default maximum row that will be saved
	// in memory for each read if input data is too large and can not be
	// consumed in one read operation.
	DefaultMaxRows = 256

	// DefDatasetMode default output mode is rows.
	DefDatasetMode = DatasetModeROWS

	// DefEOL default end-of-line
	DefEOL = '\n'
)

var (
	// ErrNoInput define an error when no Input file is given to Reader.
	ErrNoInput = errors.New("dsv: No input file is given in config")

	// ErrMissRecordsLen define an error when trying to push Row
	// to Field, when their length is not equal.
	// See reader.PushRowToColumns().
	ErrMissRecordsLen = errors.New("dsv: Mismatch between number of record in row and columns length")

	// ErrNoOutput define an error when no output file is given to Writer.
	ErrNoOutput = errors.New("dsv: No output file is given in config")

	// ErrNotOpen define an error when output file has not been opened
	// by Writer.
	ErrNotOpen = errors.New("dsv: Output file is not opened")

	// ErrNilReader define an error when Reader object is nil when passed
	// to Write function.
	ErrNilReader = errors.New("dsv: Reader object is nil")
)

// ReadWriter combine reader and writer.
type ReadWriter struct {
	Reader
	Writer
}

// New create a new ReadWriter object.
func New(config string, dataset interface{}) (rw *ReadWriter, e error) {
	rw = &ReadWriter{}

	e = rw.Reader.Init(config, dataset)
	if e != nil {
		return nil, e
	}

	e = OpenWriter(&rw.Writer, config)
	if e != nil {
		return nil, e
	}

	return
}

// SetConfigPath of input and output file.
func (dsv *ReadWriter) SetConfigPath(dir string) {
	dsv.Reader.SetConfigPath(dir)
	dsv.Writer.SetConfigPath(dir)
}

// Close reader and writer.
func (dsv *ReadWriter) Close() (e error) {
	e = dsv.Writer.Close()
	if e != nil {
		return
	}
	return dsv.Reader.Close()
}
