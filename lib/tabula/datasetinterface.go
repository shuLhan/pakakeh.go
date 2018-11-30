// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/shuLhan/share/lib/debug"
)

//
// DatasetInterface is the interface for working with DSV data.
//
type DatasetInterface interface {
	Init(mode int, types []int, names []string)
	Clone() interface{}
	Reset() error

	GetMode() int
	SetMode(mode int)

	GetNColumn() int
	GetNRow() int
	Len() int

	GetColumnsType() []int
	SetColumnsType(types []int)

	GetColumnTypeAt(idx int) (int, error)
	SetColumnTypeAt(idx, tipe int) error

	GetColumnsName() []string
	SetColumnsName(names []string)

	AddColumn(tipe int, name string, vs []string)
	GetColumn(idx int) *Column
	GetColumnByName(name string) *Column
	GetColumns() *Columns
	SetColumns(*Columns)

	GetRow(idx int) *Row
	GetRows() *Rows
	SetRows(*Rows)
	DeleteRow(idx int) *Row

	GetData() interface{}
	GetDataAsRows() *Rows
	GetDataAsColumns() *Columns

	TransposeToColumns()
	TransposeToRows()

	PushRow(r *Row)
	PushRowToColumns(r *Row)
	FillRowsWithColumn(colidx int, col Column)
	PushColumn(col Column)
	PushColumnToRows(col Column)

	MergeColumns(DatasetInterface)
	MergeRows(DatasetInterface)
}

//
// ReadDatasetConfig open dataset configuration file and initialize dataset
// field from there.
//
func ReadDatasetConfig(ds interface{}, fcfg string) (e error) {
	cfg, e := ioutil.ReadFile(fcfg)

	if nil != e {
		return e
	}

	return json.Unmarshal(cfg, ds)
}

//
// SortColumnsByIndex will sort all columns using sorted index.
//
func SortColumnsByIndex(di DatasetInterface, sortedIdx []int) {
	if di.GetMode() == DatasetModeRows {
		di.TransposeToColumns()
	}

	cols := di.GetColumns()
	for x, col := range *cols {
		colsorted := col.Records.SortByIndex(sortedIdx)
		(*cols)[x].SetRecords(colsorted)
	}
}

//
// SplitRowsByNumeric will split the data using splitVal in column `colidx`.
//
// For example, given two continuous attribute,
//
// 	A: {1,2,3,4}
// 	B: {5,6,7,8}
//
// if colidx is (1) B and splitVal is 7, the data will splitted into left set
//
// 	A': {1,2}
// 	B': {5,6}
//
// and right set
//
// 	A'': {3,4}
// 	B'': {7,8}
//
func SplitRowsByNumeric(di DatasetInterface, colidx int, splitVal float64) (
	splitLess DatasetInterface,
	splitGreater DatasetInterface,
	e error,
) {
	// check type of column
	coltype, e := di.GetColumnTypeAt(colidx)
	if e != nil {
		return nil, nil, e
	}

	if !(coltype == TInteger || coltype == TReal) {
		return nil, nil, ErrInvalidColType
	}

	// Should we convert the data mode back later.
	orgmode := di.GetMode()

	if orgmode == DatasetModeColumns {
		di.TransposeToRows()
	}

	if debug.Value >= 2 {
		fmt.Println("[tabula] dataset:", di)
	}

	splitLess = di.Clone().(DatasetInterface)
	splitGreater = di.Clone().(DatasetInterface)

	rows := di.GetRows()
	for _, row := range *rows {
		if (*row)[colidx].Float() < splitVal {
			splitLess.PushRow(row)
		} else {
			splitGreater.PushRow(row)
		}
	}

	if debug.Value >= 2 {
		fmt.Println("[tabula] split less:", splitLess)
		fmt.Println("[tabula] split greater:", splitGreater)
	}

	switch orgmode {
	case DatasetModeColumns:
		di.TransposeToColumns()
		splitLess.TransposeToColumns()
		splitGreater.TransposeToColumns()
	case DatasetModeMatrix:
		// do nothing, since its already filled when pushing new row.
	}

	return splitLess, splitGreater, nil
}

//
// SplitRowsByCategorical will split the data using a set of split value in
// column `colidx`.
//
// For example, given two attributes,
//
// 	X: [A,B,A,B,C,D,C,D]
// 	Y: [1,2,3,4,5,6,7,8]
//
// if colidx is (0) or A and split value is a set `[A,C]`, the data will
// splitted into left set which contain all rows that have A or C,
//
// 	X': [A,A,C,C]
// 	Y': [1,3,5,7]
//
// and the right set, excluded set, will contain all rows which is not A or C,
//
// 	X'': [B,B,D,D]
// 	Y'': [2,4,6,8]
//
func SplitRowsByCategorical(di DatasetInterface, colidx int, splitVal []string) (
	splitIn DatasetInterface,
	splitEx DatasetInterface,
	e error,
) {
	// check type of column
	coltype, e := di.GetColumnTypeAt(colidx)
	if e != nil {
		return nil, nil, e
	}

	if coltype != TString {
		return nil, nil, ErrInvalidColType
	}

	// should we convert the data mode back?
	orgmode := di.GetMode()

	if orgmode == DatasetModeColumns {
		di.TransposeToRows()
	}

	splitIn = di.Clone().(DatasetInterface)
	splitEx = di.Clone().(DatasetInterface)

	for _, row := range *di.GetRows() {
		found := false
		for _, val := range splitVal {
			if (*row)[colidx].String() == val {
				splitIn.PushRow(row)
				found = true
				break
			}
		}
		if !found {
			splitEx.PushRow(row)
		}
	}

	// convert all dataset based on original
	switch orgmode {
	case DatasetModeColumns:
		di.TransposeToColumns()
		splitIn.TransposeToColumns()
		splitEx.TransposeToColumns()
	case DatasetModeMatrix, DatasetNoMode:
		splitIn.TransposeToColumns()
		splitEx.TransposeToColumns()
	}

	return splitIn, splitEx, nil
}

//
// SplitRowsByValue generic function to split data by value. This function will
// split data using value in column `colidx`. If value is numeric it will return
// any rows that have column value less than `value` in `splitL`, and any column
// value greater or equal to `value` in `splitR`.
//
func SplitRowsByValue(di DatasetInterface, colidx int, value interface{}) (
	splitL DatasetInterface,
	splitR DatasetInterface,
	e error,
) {
	coltype, e := di.GetColumnTypeAt(colidx)
	if e != nil {
		return nil, nil, e
	}

	if coltype == TString {
		splitL, splitR, e = SplitRowsByCategorical(di, colidx,
			value.([]string))
	} else {
		var splitval float64

		switch v := value.(type) {
		case int:
			splitval = float64(v)
		case int64:
			splitval = float64(v)
		case float32:
			splitval = float64(v)
		case float64:
			splitval = v
		}

		splitL, splitR, e = SplitRowsByNumeric(di, colidx, splitval)
	}

	if e != nil {
		return nil, nil, e
	}

	return splitL, splitR, nil
}

//
// SelectRowsWhere return all rows which column value in `colidx` is equal to
// `colval`.
//
func SelectRowsWhere(dataset DatasetInterface, colidx int, colval string) DatasetInterface {
	orgmode := dataset.GetMode()

	if orgmode == DatasetModeColumns {
		dataset.TransposeToRows()
	}

	selected := NewDataset(dataset.GetMode(), nil, nil)

	selected.Rows = dataset.GetRows().SelectWhere(colidx, colval)

	switch orgmode {
	case DatasetModeColumns:
		dataset.TransposeToColumns()
		selected.TransposeToColumns()
	case DatasetModeMatrix, DatasetNoMode:
		selected.TransposeToColumns()
	}

	return selected
}

//
// RandomPickRows return `n` item of row that has been selected randomly from
// dataset.Rows. The ids of rows that has been picked is saved id `pickedIdx`.
//
// If duplicate is true, the row that has been picked can be picked up again,
// otherwise it only allow one pick. This is also called as random selection
// with or without replacement in machine learning domain.
//
// If output mode is columns, it will be transposed to rows.
//
func RandomPickRows(dataset DatasetInterface, n int, duplicate bool) (
	picked DatasetInterface,
	unpicked DatasetInterface,
	pickedIdx []int,
	unpickedIdx []int,
) {
	orgmode := dataset.GetMode()

	if orgmode == DatasetModeColumns {
		dataset.TransposeToRows()
	}

	picked = dataset.Clone().(DatasetInterface)
	unpicked = dataset.Clone().(DatasetInterface)

	pickedRows, unpickedRows, pickedIdx, unpickedIdx :=
		dataset.GetRows().RandomPick(n, duplicate)

	picked.SetRows(&pickedRows)
	unpicked.SetRows(&unpickedRows)

	// switch the dataset based on original mode
	switch orgmode {
	case DatasetModeColumns:
		dataset.TransposeToColumns()
		// transform the picked and unpicked set.
		picked.TransposeToColumns()
		unpicked.TransposeToColumns()

	case DatasetModeMatrix, DatasetNoMode:
		// transform the picked and unpicked set.
		picked.TransposeToColumns()
		unpicked.TransposeToColumns()
	}

	return picked, unpicked, pickedIdx, unpickedIdx
}

//
// RandomPickColumns will select `n` column randomly from dataset and return
// new dataset with picked and unpicked columns, and their column index.
//
// If duplicate is true, column that has been pick up can be pick up again.
//
// If dataset output mode is rows, it will transposed to columns.
//
func RandomPickColumns(dataset DatasetInterface, n int, dup bool, excludeIdx []int) (
	picked DatasetInterface,
	unpicked DatasetInterface,
	pickedIdx []int,
	unpickedIdx []int,
) {
	orgmode := dataset.GetMode()

	if orgmode == DatasetModeRows {
		dataset.TransposeToColumns()
	}

	picked = dataset.Clone().(DatasetInterface)
	unpicked = dataset.Clone().(DatasetInterface)

	pickedColumns, unpickedColumns, pickedIdx, unpickedIdx :=
		dataset.GetColumns().RandomPick(n, dup, excludeIdx)

	picked.SetColumns(&pickedColumns)
	unpicked.SetColumns(&unpickedColumns)

	// transpose picked and unpicked dataset based on original mode
	switch orgmode {
	case DatasetModeRows:
		dataset.TransposeToRows()
		picked.TransposeToRows()
		unpicked.TransposeToRows()
	case DatasetModeMatrix, DatasetNoMode:
		picked.TransposeToRows()
		unpicked.TransposeToRows()
	}

	return picked, unpicked, pickedIdx, unpickedIdx
}

//
// SelectColumnsByIdx return new dataset with selected column index.
//
func SelectColumnsByIdx(dataset DatasetInterface, colsIdx []int) (
	newset DatasetInterface,
) {
	var col *Column

	orgmode := dataset.GetMode()

	if orgmode == DatasetModeRows {
		dataset.TransposeToColumns()
	}

	newset = dataset.Clone().(DatasetInterface)

	for _, idx := range colsIdx {
		col = dataset.GetColumn(idx)
		if col == nil {
			continue
		}

		newset.PushColumn(*col)
	}

	// revert the mode back
	switch orgmode {
	case DatasetModeRows:
		dataset.TransposeToRows()
		newset.TransposeToRows()
	case DatasetModeColumns:
		// do nothing
	case DatasetModeMatrix:
		// do nothing
	}

	return newset
}
