// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"testing"

	"github.com/shuLhan/share/lib/tabula"
)

//
// TestWriter test reading and writing DSV.
//
func TestWriter(t *testing.T) {
	rw, e := New("testdata/config.dsv", nil)
	if e != nil {
		t.Fatal(e)
	}

	doReadWrite(t, &rw.Reader, &rw.Writer, expectation, true)

	e = rw.Close()
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, rw.GetOutput(), "testdata/expected.dat", true)
}

//
// TestWriterWithSkip test reading and writing DSV with some column in input being
// skipped.
//
func TestWriterWithSkip(t *testing.T) {
	rw, e := New("testdata/config_skip.dsv", nil)
	if e != nil {
		t.Fatal(e)
	}

	doReadWrite(t, &rw.Reader, &rw.Writer, expSkip, true)

	e = rw.Close()
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, rw.GetOutput(), "testdata/expected_skip.dat", true)
}

//
// TestWriterWithColumns test reading and writing DSV with where each row
// is saved in DatasetMode = 'columns'.
//
func TestWriterWithColumns(t *testing.T) {
	rw, e := New("testdata/config_skip.dsv", nil)
	if e != nil {
		t.Fatal(e)
	}

	rw.SetDatasetMode(DatasetModeCOLUMNS)

	doReadWrite(t, &rw.Reader, &rw.Writer, expSkipColumns, true)

	e = rw.Close()
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, "testdata/expected_skip.dat", rw.GetOutput(), true)
}

func TestWriteRawRows(t *testing.T) {
	dataset := tabula.NewDataset(tabula.DatasetModeRows, nil, nil)

	populateWithRows(t, dataset)

	writer, e := NewWriter("")
	if e != nil {
		t.Fatal(e)
	}

	outfile := "testdata/writerawrows.out"
	expfile := "testdata/writeraw.exp"

	e = writer.OpenOutput(outfile)
	if e != nil {
		t.Fatal(e)
	}

	_, e = writer.WriteRawDataset(dataset, nil)
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, outfile, expfile, true)
}

func TestWriteRawColumns(t *testing.T) {
	var e error

	dataset := tabula.NewDataset(tabula.DatasetModeColumns, nil, nil)

	populateWithColumns(t, dataset)

	writer, e := NewWriter("")
	if e != nil {
		t.Fatal(e)
	}

	outfile := "testdata/writerawcolumns.out"
	expfile := "testdata/writeraw.exp"

	e = writer.OpenOutput(outfile)
	if e != nil {
		t.Fatal(e)
	}

	_, e = writer.WriteRawDataset(dataset, nil)
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, outfile, expfile, true)
}
