// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/tabula"
	"github.com/shuLhan/share/lib/test"
)

//nolint:gochecknoglobals
var jsonSample = []string{
	`{}`,
	`{
		"Input"		:"testdata/input.dat"
	}`,
	`{
		"Input"		:"testdata/input.dat"
	}`,
	`{
		"Input"		:"testdata/input.dat"
	,	"InputMetadata"	:
		[{
			"Name"		:"A"
		,	"Separator"	:","
		},{
			"Name"		:"B"
		,	"Separator"	:";"
		}]
	}`,
	`{
		"Input"		:"testdata/input.dat"
	,	"Skip"		:1
	,	"MaxRows"	:1
	,	"InputMetadata"	:
		[{
			"Name"		:"id"
		,	"Separator"	:";"
		,	"Type"		:"integer"
		},{
			"Name"		:"name"
		,	"Separator"	:"-"
		,	"LeftQuote"	:"\""
		,	"RightQuote"	:"\""
		},{
			"Name"		:"value"
		,	"Separator"	:";"
		,	"LeftQuote"	:"[["
		,	"RightQuote"	:"]]"
		},{
			"Name"		:"integer"
		,	"Type"		:"integer"
		,	"Separator"	:";"
		},{
			"Name"		:"real"
		,	"Type"		:"real"
		}]
	}`,
	`{
		"Input"		:"testdata/input.dat"
	,	"Skip"		:1
	,	"MaxRows"	:1
	,	"InputMetadata"	:
		[{
			"Name"		:"id"
		},{
			"Name"		:"editor"
		},{
			"Name"		:"old_rev_id"
		},{
			"Name"		:"new_rev_id"
		},{
			"Name"		:"diff_url"
		},{
			"Name"		:"edit_time"
		},{
			"Name"		:"edit_comment"
		},{
			"Name"		:"article_id"
		},{
			"Name"		:"article_title"
		}]
	}`,
}

//nolint:gochecknoglobals
var readers = []*Reader{
	{},
	{
		Input: "testdata/input.dat",
	},
	{
		Input: "test-another.dsv",
	},
	{
		Input: "testdata/input.dat",
		InputMetadata: []Metadata{
			{
				Name:      "A",
				Separator: ",",
			},
			{
				Name:      "B",
				Separator: ";",
			},
		},
	},
}

//
// TestReaderNoInput will print error that the input is not defined.
//
func TestReaderNoInput(t *testing.T) {
	dsvReader := &Reader{}

	e := ConfigParse(dsvReader, []byte(jsonSample[0]))

	if nil != e {
		t.Fatal(e)
	}

	e = dsvReader.Init("", nil)

	if nil == e {
		t.Fatal("TestReaderNoInput: failed, should return non nil!")
	}
}

//
// TestConfigParse test parsing metadata.
//
func TestConfigParse(t *testing.T) {
	cases := []struct {
		in  string
		out *Reader
	}{
		{
			jsonSample[1],
			readers[1],
		},
		{
			jsonSample[3],
			readers[3],
		},
	}

	dsvReader := &Reader{}

	for _, c := range cases {
		e := ConfigParse(dsvReader, []byte(c.in))

		if e != nil {
			t.Fatal(e)
		}
		if !dsvReader.IsEqual(c.out) {
			t.Fatal("Test failed on ", c.in)
		}
	}
}

func TestReaderIsEqual(t *testing.T) {
	cases := []struct {
		in     *Reader
		out    *Reader
		result bool
	}{
		{
			readers[1],
			&Reader{
				Input: "testdata/input.dat",
			},
			true,
		},
		{
			readers[1],
			readers[2],
			false,
		},
	}

	var r bool

	for _, c := range cases {
		r = c.in.IsEqual(c.out)

		if r != c.result {
			t.Fatal("Test failed on equality between ", c.in,
				"\n and ", c.out)
		}
	}
}

//
// doRead test reading the DSV data.
//
func doRead(t *testing.T, dsvReader *Reader, exp []string) {
	i := 0
	var n int
	var e error

	for {
		n, e = Read(dsvReader)

		if n > 0 {
			r := fmt.Sprint(dsvReader.
				GetDataset().(tabula.DatasetInterface).
				GetDataAsRows())

			test.Assert(t, "", exp[i], r, true)

			i++
		} else if e == io.EOF {
			// EOF
			break
		}
	}
}

//
// TestReader test reading.
//
func TestReaderRead(t *testing.T) {
	dsvReader := &Reader{}

	e := ConfigParse(dsvReader, []byte(jsonSample[4]))

	if nil != e {
		t.Fatal(e)
	}

	e = dsvReader.Init("", nil)
	if nil != e {
		t.Fatal(e)
	}

	doRead(t, dsvReader, expectation)

	e = dsvReader.Close()
	if e != nil {
		t.Fatal(e)
	}
}

//
// TestReaderOpen real example from the start.
//
func TestReaderOpen(t *testing.T) {
	dsvReader, e := NewReader("testdata/config.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	doRead(t, dsvReader, expectation)

	e = dsvReader.Close()
	if e != nil {
		t.Fatal(e)
	}
}

func TestDatasetMode(t *testing.T) {
	var e error
	var config = []string{`{
		"Input"		:"testdata/input.dat"
	,	"DatasetMode"	:"row"
	}`, `{
		"Input"		:"testdata/input.dat"
	,	"DatasetMode"	:"rows"
	}`, `{
		"Input"		:"testdata/input.dat"
	,	"DatasetMode"	:"columns"
	}`}

	var exps = []struct {
		status bool
		value  string
	}{{
		status: false,
		value:  config[0],
	}, {
		status: true,
		value:  config[1],
	}, {
		status: true,
		value:  config[2],
	}}

	reader := &Reader{}

	for k, v := range exps {
		e = ConfigParse(reader, []byte(config[k]))

		if e != nil {
			t.Fatal(e)
		}

		e = reader.Init("", nil)
		if e != nil {
			if v.status {
				t.Fatal(e)
			}
		}
	}
}

func TestReaderToColumns(t *testing.T) {
	reader := &Reader{}

	e := ConfigParse(reader, []byte(jsonSample[4]))
	if nil != e {
		t.Fatal(e)
	}

	e = reader.Init("", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader.SetDatasetMode(DatasetModeCOLUMNS)

	var n, i int
	for {
		n, e = Read(reader)

		if n > 0 {
			ds := reader.GetDataset().(tabula.DatasetInterface)
			ds.TransposeToRows()

			r := fmt.Sprint(ds.GetData())

			test.Assert(t, "", expectation[i], r, true)

			i++
		} else if e == io.EOF {
			// EOF
			break
		}
	}
}

//
// TestReaderSkip will test the 'Skip' option in Metadata.
//
func TestReaderSkip(t *testing.T) {
	dsvReader, e := NewReader("testdata/config_skip.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	doRead(t, dsvReader, expSkip)

	e = dsvReader.Close()
	if e != nil {
		t.Fatal(e)
	}
}

func TestTransposeToColumns(t *testing.T) {
	reader, e := NewReader("testdata/config_skip.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader.SetMaxRows(-1)

	_, e = Read(reader)

	if e != io.EOF {
		t.Fatal(e)
	}

	ds := reader.GetDataset().(tabula.DatasetInterface)
	ds.TransposeToColumns()

	exp := fmt.Sprint(expSkipColumnsAll)

	columns := ds.GetDataAsColumns()

	got := fmt.Sprint(*columns)

	test.Assert(t, "", exp, got, true)

	e = reader.Close()
	if e != nil {
		t.Fatal(e)
	}
}

func TestSortColumnsByIndex(t *testing.T) {
	reader, e := NewReader("testdata/config_skip.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader.SetMaxRows(-1)

	_, e = Read(reader)
	if e != io.EOF {
		t.Fatal(e)
	}

	// reverse the data
	var idxReverse []int
	var expReverse []string

	for x := len(expSkip) - 1; x >= 0; x-- {
		idxReverse = append(idxReverse, x)
		expReverse = append(expReverse, expSkip[x])
	}

	ds := reader.GetDataset().(tabula.DatasetInterface)

	tabula.SortColumnsByIndex(ds, idxReverse)

	exp := strings.Join(expReverse, "")
	got := fmt.Sprint(ds.GetDataAsRows())

	test.Assert(t, "", exp, got, true)

	exp = "[" + strings.Join(expSkipColumnsAllRev, " ") + "]"

	columns := ds.GetDataAsColumns()

	got = fmt.Sprint(*columns)

	test.Assert(t, "", exp, got, true)

	e = reader.Close()
	if e != nil {
		t.Fatal(e)
	}
}

func TestSplitRowsByValue(t *testing.T) {
	reader, e := NewReader("testdata/config.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader.SetMaxRows(256)

	_, e = Read(reader)

	if e != nil && e != io.EOF {
		t.Fatal(e)
	}

	ds := reader.GetDataset().(tabula.DatasetInterface)
	splitL, splitR, e := tabula.SplitRowsByValue(ds, 0, 6)

	if e != nil {
		t.Fatal(e)
	}

	// test left split
	exp := ""
	for x := 0; x < 4; x++ {
		exp += expectation[x]
	}

	got := fmt.Sprint(splitL.GetDataAsRows())

	test.Assert(t, "", exp, got, true)

	// test right split
	exp = ""
	for x := 4; x < len(expectation); x++ {
		exp += expectation[x]
	}

	got = fmt.Sprint(splitR.GetDataAsRows())

	test.Assert(t, "", exp, got, true)

	e = reader.Close()
	if e != nil {
		t.Fatal(e)
	}
}

//
// testWriteOutput will write merged reader and check with expected file output.
//
func testWriteOutput(t *testing.T, r *Reader, outfile, expfile string) {

	writer, e := NewWriter("")
	if e != nil {
		t.Fatal(e)
	}

	e = writer.OpenOutput(outfile)

	if e != nil {
		t.Fatal(e)
	}

	sep := "\t"
	ds := r.GetDataset().(tabula.DatasetInterface)

	_, e = writer.WriteRawDataset(ds, &sep)
	if e != nil {
		t.Fatal(e)
	}

	e = writer.Close()
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, outfile, expfile)
}

func TestMergeColumns(t *testing.T) {
	reader1, e := NewReader("testdata/config.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader2, e := NewReader("testdata/config_skip.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader1.SetMaxRows(-1)
	reader2.SetMaxRows(-1)

	_, e = Read(reader1)
	if e != io.EOF {
		t.Fatal(e)
	}

	_, e = Read(reader2)
	if e != io.EOF {
		t.Fatal(e)
	}

	e = reader1.Close()
	if e != nil {
		t.Fatal(e)
	}

	e = reader2.Close()
	if e != nil {
		t.Fatal(e)
	}

	reader1.InputMetadata[len(reader1.InputMetadata)-1].Separator = ";"

	reader1.MergeColumns(reader2)

	outfile := "testdata/output_merge_columns.dat"
	expfile := "testdata/expected_merge_columns.dat"

	testWriteOutput(t, reader1, outfile, expfile)
}

func TestMergeRows(t *testing.T) {
	reader1, e := NewReader("testdata/config.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader2, e := NewReader("testdata/config_skip.dsv", nil)
	if nil != e {
		t.Fatal(e)
	}

	reader1.SetMaxRows(-1)
	reader2.SetMaxRows(-1)

	_, e = Read(reader1)
	if e != io.EOF {
		t.Fatal(e)
	}

	_, e = Read(reader2)
	if e != io.EOF {
		t.Fatal(e)
	}

	e = reader1.Close()
	if e != nil {
		t.Fatal(e)
	}

	e = reader2.Close()
	if e != nil {
		t.Fatal(e)
	}

	reader1.MergeRows(reader2)

	outfile := "testdata/output_merge_rows.dat"
	expfile := "testdata/expected_merge_rows.dat"

	testWriteOutput(t, reader1, outfile, expfile)
}
