// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"runtime/debug"
	"testing"

	"github.com/shuLhan/share/lib/tabula"
	"github.com/shuLhan/share/lib/test"
)

//
// assertFile compare content of two file, print error message and exit
// when both are different.
//
func assertFile(t *testing.T, a, b string) {
	out, e := ioutil.ReadFile(a)

	if nil != e {
		debug.PrintStack()
		t.Error(e)
	}

	exp, e := ioutil.ReadFile(b)

	if nil != e {
		debug.PrintStack()
		t.Error(e)
	}

	r := bytes.Compare(out, exp)

	if 0 != r {
		debug.PrintStack()
		t.Fatal("Comparing", a, "with", b, ": result is different (",
			r, ")")
	}
}

func checkDataset(t *testing.T, r *Reader, exp string) {
	var got string
	ds := r.GetDataset().(tabula.DatasetInterface)
	data := ds.GetData()

	switch v := data.(type) {
	case *tabula.Rows:
		rows := v
		got = fmt.Sprint(*rows)
	case *tabula.Columns:
		cols := v
		got = fmt.Sprint(*cols)
	case *tabula.Matrix:
		matrix := v
		got = fmt.Sprint(*matrix)
	default:
		fmt.Println("data type unknown")
	}

	test.Assert(t, "", exp, got, true)
}

//
// doReadWrite test reading and writing the DSV data.
//
func doReadWrite(t *testing.T, dsvReader *Reader, dsvWriter *Writer,
	expectation []string, check bool) {
	i := 0

	for {
		n, e := Read(dsvReader)

		if e == io.EOF || n == 0 {
			_, e = dsvWriter.Write(dsvReader)
			if e != nil {
				t.Fatal(e)
			}

			break
		}

		if e != nil {
			continue
		}

		if check {
			checkDataset(t, dsvReader, expectation[i])
			i++
		}

		_, e = dsvWriter.Write(dsvReader)
		if e != nil {
			t.Fatal(e)
		}
	}

	e := dsvWriter.Flush()
	if e != nil {
		t.Fatal(e)
	}
}

var datasetRows = [][]string{ // nolint: gochecknoglobals
	{"0", "1", "A"},
	{"1", "1.1", "B"},
	{"2", "1.2", "A"},
	{"3", "1.3", "B"},
	{"4", "1.4", "C"},
	{"5", "1.5", "D"},
	{"6", "1.6", "C"},
	{"7", "1.7", "D"},
	{"8", "1.8", "E"},
	{"9", "1.9", "F"},
}

var datasetCols = [][]string{ // nolint: gochecknoglobals
	{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
	{"1", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9"},
	{"A", "B", "A", "B", "C", "D", "C", "D", "E", "F"},
}

var datasetTypes = []int{ // nolint: gochecknoglobals
	tabula.TInteger,
	tabula.TReal,
	tabula.TString,
}

var datasetNames = []string{"int", "real", "string"} // nolint: gochecknoglobals

func populateWithRows(t *testing.T, dataset *tabula.Dataset) {
	for _, rowin := range datasetRows {
		row := make(tabula.Row, len(rowin))

		for x, recin := range rowin {
			rec, e := tabula.NewRecordBy(recin, datasetTypes[x])
			if e != nil {
				t.Fatal(e)
			}

			row[x] = rec
		}

		dataset.PushRow(&row)
	}
}

func populateWithColumns(t *testing.T, dataset *tabula.Dataset) {
	for x := range datasetCols {
		col, e := tabula.NewColumnString(datasetCols[x], datasetTypes[x],
			datasetNames[x])
		if e != nil {
			t.Fatal(e)
		}

		dataset.PushColumn(*col)
	}
}
