// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

//nolint:gochecknoglobals
var (
	datasetRows = [][]string{
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

	datasetCols = [][]string{
		{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		{"1", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9"},
		{"A", "B", "A", "B", "C", "D", "C", "D", "E", "F"},
	}

	datasetTypes = []int{
		TInteger,
		TReal,
		TString,
	}

	datasetNames = []string{"int", "real", "string"}
)

func populateWithRows(dataset *Dataset) error {
	for _, rowin := range datasetRows {
		row := make(Row, len(rowin))

		for x, recin := range rowin {
			rec, e := NewRecordBy(recin, datasetTypes[x])
			if e != nil {
				return e
			}

			row[x] = rec
		}

		dataset.PushRow(&row)
	}
	return nil
}

func populateWithColumns(t *testing.T, dataset *Dataset) {
	for x := range datasetCols {
		col, e := NewColumnString(datasetCols[x], datasetTypes[x],
			datasetNames[x])
		if e != nil {
			t.Fatal(e)
		}

		dataset.PushColumn(*col)
	}
}

func createDataset(t *testing.T) (dataset *Dataset) {
	dataset = NewDataset(DatasetModeRows, datasetTypes,
		datasetNames)

	e := populateWithRows(dataset)
	if e != nil {
		t.Fatal(e)
	}

	return
}

func DatasetStringJoinByIndex(t *testing.T, dataset [][]string, indis []int) (res string) {
	for x := range indis {
		res += fmt.Sprint("&", dataset[indis[x]])
	}
	return res
}

func DatasetRowsJoin(t *testing.T) (s string) {
	for x := range datasetRows {
		s += fmt.Sprint("&", datasetRows[x])
	}
	return
}

func DatasetColumnsJoin(t *testing.T) (s string) {
	for x := range datasetCols {
		s += fmt.Sprint(datasetCols[x])
	}
	return
}

func TestSplitRowsByNumeric(t *testing.T) {
	dataset := createDataset(t)

	// Split integer by float
	splitL, splitR, e := SplitRowsByNumeric(dataset, 0, 4.5)
	if e != nil {
		t.Fatal(e)
	}

	expIdx := []int{0, 1, 2, 3, 4}
	exp := DatasetStringJoinByIndex(t, datasetRows, expIdx)
	rows := splitL.GetDataAsRows()
	got := fmt.Sprint(rows)

	test.Assert(t, "", exp, got, true)

	expIdx = []int{5, 6, 7, 8, 9}
	exp = DatasetStringJoinByIndex(t, datasetRows, expIdx)
	got = fmt.Sprint(splitR.GetDataAsRows())

	test.Assert(t, "", exp, got, true)

	// Split by float
	splitL, splitR, e = SplitRowsByNumeric(dataset, 1, 1.8)
	if e != nil {
		t.Fatal(e)
	}

	expIdx = []int{0, 1, 2, 3, 4, 5, 6, 7}
	exp = DatasetStringJoinByIndex(t, datasetRows, expIdx)
	got = fmt.Sprint(splitL.GetDataAsRows())

	test.Assert(t, "", exp, got, true)

	expIdx = []int{8, 9}
	exp = DatasetStringJoinByIndex(t, datasetRows, expIdx)
	got = fmt.Sprint(splitR.GetDataAsRows())

	test.Assert(t, "", exp, got, true)
}

func TestSplitRowsByCategorical(t *testing.T) {
	dataset := createDataset(t)
	splitval := []string{"A", "D"}

	splitL, splitR, e := SplitRowsByCategorical(dataset, 2,
		splitval)
	if e != nil {
		t.Fatal(e)
	}

	expIdx := []int{0, 2, 5, 7}
	exp := DatasetStringJoinByIndex(t, datasetRows, expIdx)
	got := fmt.Sprint(splitL.GetDataAsRows())

	test.Assert(t, "", exp, got, true)

	expIdx = []int{1, 3, 4, 6, 8, 9}
	exp = DatasetStringJoinByIndex(t, datasetRows, expIdx)
	got = fmt.Sprint(splitR.GetDataAsRows())

	test.Assert(t, "", exp, got, true)
}

func TestModeColumnsPushColumn(t *testing.T) {
	dataset := NewDataset(DatasetModeColumns, nil, nil)

	exp := ""
	got := ""
	for x := range datasetCols {
		col, e := NewColumnString(datasetCols[x], datasetTypes[x],
			datasetNames[x])
		if e != nil {
			t.Fatal(e)
		}

		dataset.PushColumn(*col)

		exp += fmt.Sprint(datasetCols[x])
		got += fmt.Sprint(dataset.Columns[x].Records)
	}

	test.Assert(t, "", exp, got, true)

	// Check rows
	exp = ""
	got = fmt.Sprint(dataset.Rows)
	test.Assert(t, "", exp, got, true)
}

func TestModeRowsPushColumn(t *testing.T) {
	dataset := NewDataset(DatasetModeRows, nil, nil)

	populateWithColumns(t, dataset)

	// Check rows
	exp := DatasetRowsJoin(t)
	got := fmt.Sprint(dataset.Rows)

	test.Assert(t, "", exp, got, true)

	// Check columns
	exp = "[{int 1 0 [] []} {real 2 0 [] []} {string 0 0 [] []}]"
	got = fmt.Sprint(dataset.Columns)

	test.Assert(t, "", exp, got, true)
}

func TestModeMatrixPushColumn(t *testing.T) {
	dataset := NewDataset(DatasetModeMatrix, nil, nil)

	exp := ""
	got := ""
	for x := range datasetCols {
		col, e := NewColumnString(datasetCols[x], datasetTypes[x],
			datasetNames[x])
		if e != nil {
			t.Fatal(e)
		}

		dataset.PushColumn(*col)

		exp += fmt.Sprint(datasetCols[x])
		got += fmt.Sprint(dataset.Columns[x].Records)
	}

	test.Assert(t, "", exp, got, true)

	// Check rows
	exp = DatasetRowsJoin(t)
	got = fmt.Sprint(dataset.Rows)

	test.Assert(t, "", exp, got, true)
}

func TestModeRowsPushRows(t *testing.T) {
	dataset := NewDataset(DatasetModeRows, nil, nil)

	e := populateWithRows(dataset)
	if e != nil {
		t.Fatal(e)
	}

	exp := DatasetRowsJoin(t)
	got := fmt.Sprint(dataset.Rows)

	test.Assert(t, "", exp, got, true)
}

func TestModeColumnsPushRows(t *testing.T) {
	dataset := NewDataset(DatasetModeColumns, nil, nil)

	e := populateWithRows(dataset)
	if e != nil {
		t.Fatal(e)
	}

	// check rows
	exp := ""
	got := fmt.Sprint(dataset.Rows)

	test.Assert(t, "", exp, got, true)

	// check columns
	exp = DatasetColumnsJoin(t)
	got = ""
	for x := range dataset.Columns {
		got += fmt.Sprint(dataset.Columns[x].Records)
	}

	test.Assert(t, "", exp, got, true)
}

func TestModeMatrixPushRows(t *testing.T) {
	dataset := NewDataset(DatasetModeMatrix, nil, nil)

	e := populateWithRows(dataset)
	if e != nil {
		t.Fatal(e)
	}

	exp := DatasetRowsJoin(t)
	got := fmt.Sprint(dataset.Rows)

	test.Assert(t, "", exp, got, true)

	// check columns
	exp = DatasetColumnsJoin(t)
	got = ""
	for x := range dataset.Columns {
		got += fmt.Sprint(dataset.Columns[x].Records)
	}

	test.Assert(t, "", exp, got, true)
}

func TestSelectRowsWhere(t *testing.T) {
	dataset := NewDataset(DatasetModeMatrix, nil, nil)

	e := populateWithRows(dataset)
	if e != nil {
		t.Fatal(e)
	}

	// select all rows where the first column value is 9.
	selected := SelectRowsWhere(dataset, 0, "9")
	exp := dataset.GetRow(9)
	got := selected.GetRow(0)

	test.Assert(t, "", exp, got, true)
}

func TestDeleteRow(t *testing.T) {
	dataset := NewDataset(DatasetModeMatrix, nil, nil)

	e := populateWithRows(dataset)
	if e != nil {
		t.Fatal(e)
	}

	delIdx := 2

	// Check rows len.
	exp := dataset.Len() - 1
	dataset.DeleteRow(delIdx)
	got := dataset.Len()

	test.Assert(t, "", exp, got, true)

	// Check columns len.
	for _, col := range dataset.Columns {
		got = col.Len()

		test.Assert(t, "", exp, got, true)
	}

	// Check rows data.
	ridx := 0
	for x, row := range datasetRows {
		if x == delIdx {
			continue
		}
		exp := fmt.Sprint("&", row)
		got := fmt.Sprint(dataset.GetRow(ridx))
		ridx++

		test.Assert(t, "", exp, got, true)
	}

	// Check columns data.
	for x := range dataset.Columns {
		col := datasetCols[x]

		coldel := []string{}
		coldel = append(coldel, col[:delIdx]...)
		coldel = append(coldel, col[delIdx+1:]...)

		exp := fmt.Sprint(coldel)
		got := fmt.Sprint(dataset.Columns[x].Records)
		test.Assert(t, "", exp, got, true)
	}
}
