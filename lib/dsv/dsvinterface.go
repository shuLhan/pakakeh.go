// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"io"
)

//
// SimpleRead provide a shortcut to read data from file using configuration file
// from `fcfg`.
// Return the reader contained data or error if failed.
// Reader object upon returned has been closed, so if one need to read all
// data in it simply set the `MaxRows` to `-1` in config file.
//
func SimpleRead(fcfg string, dataset interface{}) (
	reader ReaderInterface,
	e error,
) {
	reader, e = NewReader(fcfg, dataset)

	if e != nil {
		return
	}

	_, e = Read(reader)
	if e != nil && e != io.EOF {
		return nil, e
	}

	e = reader.Close()

	return
}

//
// SimpleWrite provide a shortcut to write data from reader using output metadata
// format and output file defined in file `fcfg`.
//
func SimpleWrite(reader ReaderInterface, fcfg string) (nrows int, e error) {
	writer, e := NewWriter(fcfg)
	if e != nil {
		return
	}

	nrows, e = writer.Write(reader)
	if e != nil {
		return
	}

	e = writer.Close()

	return
}

//
// SimpleMerge provide a shortcut to merge two dsv files using configuration
// files passed in parameters.
//
// One must remember to set,
// - "MaxRows" to -1 to be able to read all rows, in both input configuration, and
// - "DatasetMode" to "columns" to speeding up process.
//
// This function return the merged reader or error if failed.
//
func SimpleMerge(fin1, fin2 string, dataset1, dataset2 interface{}) (
	ReaderInterface,
	error,
) {
	reader1, e := SimpleRead(fin1, dataset1)
	if e != nil {
		return nil, e
	}

	reader2, e := SimpleRead(fin2, dataset2)
	if e != nil {
		return nil, e
	}

	reader1.MergeColumns(reader2)

	return reader1, nil
}
