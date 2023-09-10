// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"errors"
	"math"
)

const (
	// DatasetNoMode default to matrix.
	DatasetNoMode = 0
	// DatasetModeRows for output mode in rows.
	DatasetModeRows = 1
	// DatasetModeColumns for output mode in columns.
	DatasetModeColumns = 2
	// DatasetModeMatrix will save data in rows and columns.
	DatasetModeMatrix = 4
)

var (
	// ErrColIdxOutOfRange operation on column index is invalid
	ErrColIdxOutOfRange = errors.New("tabula: Column index out of range")
	// ErrInvalidColType operation on column with different type
	ErrInvalidColType = errors.New("tabula: Invalid column type")
	// ErrMisColLength returned when operation on columns does not match
	// between parameter and their length
	ErrMisColLength = errors.New("tabula: mismatch on column length")
)

// Dataset contain the data, mode of saved data, number of columns and rows in
// data.
type Dataset struct {
	// Columns is input data that has been parsed.
	Columns Columns

	// Rows is input data that has been parsed.
	Rows Rows

	// Mode define the numeric value of output mode.
	Mode int
}

// NewDataset create new dataset, use the mode to initialize the dataset.
func NewDataset(mode int, types []int, names []string) (
	dataset *Dataset,
) {
	dataset = &Dataset{}

	dataset.Init(mode, types, names)

	return
}

// Init will set the dataset using mode and types.
func (dataset *Dataset) Init(mode int, types []int, names []string) {
	if types == nil {
		dataset.Columns = make(Columns, 0)
	} else {
		dataset.Columns = make(Columns, len(types))
		dataset.Columns.SetTypes(types)
	}

	dataset.SetColumnsName(names)
	dataset.SetMode(mode)
}

// Clone return a copy of current dataset.
func (dataset *Dataset) Clone() interface{} {
	clone := NewDataset(dataset.GetMode(), nil, nil)

	for _, col := range dataset.Columns {
		newcol := Column{
			Type:       col.Type,
			Name:       col.Name,
			ValueSpace: col.ValueSpace,
		}
		clone.PushColumn(newcol)
	}

	return clone
}

// Reset all data and attributes.
func (dataset *Dataset) Reset() error {
	dataset.Rows = Rows{}
	dataset.Columns.Reset()
	return nil
}

// GetMode return mode of data.
func (dataset *Dataset) GetMode() int {
	return dataset.Mode
}

// SetMode of saved data to `mode`.
func (dataset *Dataset) SetMode(mode int) {
	switch mode {
	case DatasetModeRows:
		dataset.Mode = DatasetModeRows
		dataset.Rows = make(Rows, 0)
	case DatasetModeColumns:
		dataset.Mode = DatasetModeColumns
		dataset.Columns.Reset()
	default:
		dataset.Mode = DatasetModeMatrix
		dataset.Rows = make(Rows, 0)
		dataset.Columns.Reset()
	}
}

// GetNColumn return the number of column in dataset.
func (dataset *Dataset) GetNColumn() (ncol int) {
	ncol = len(dataset.Columns)

	if ncol > 0 {
		return
	}

	if dataset.Mode == DatasetModeRows {
		if len(dataset.Rows) == 0 {
			return 0
		}
		return dataset.Rows[0].Len()
	}

	return
}

// GetNRow return number of rows in dataset.
func (dataset *Dataset) GetNRow() (nrow int) {
	switch dataset.Mode {
	case DatasetModeRows:
		nrow = len(dataset.Rows)
	case DatasetModeColumns:
		if len(dataset.Columns) == 0 {
			nrow = 0
		} else {
			// get length of record in the first column
			nrow = dataset.Columns[0].Len()
		}
	case DatasetModeMatrix, DatasetNoMode:
		// matrix mode could have empty either in rows or column.
		nrow = len(dataset.Rows)
	}
	return
}

// Len return number of row in dataset.
func (dataset *Dataset) Len() int {
	return dataset.GetNRow()
}

// GetColumnsType return the type of all columns.
func (dataset *Dataset) GetColumnsType() (types []int) {
	for x := range dataset.Columns {
		types = append(types, dataset.Columns[x].Type)
	}

	return
}

// SetColumnsType of data in all columns.
func (dataset *Dataset) SetColumnsType(types []int) {
	dataset.Columns = make(Columns, len(types))
	dataset.Columns.SetTypes(types)
}

// GetColumnTypeAt return type of column in index `colidx` in dataset.
func (dataset *Dataset) GetColumnTypeAt(idx int) (int, error) {
	if idx >= dataset.GetNColumn() {
		return TUndefined, ErrColIdxOutOfRange
	}

	return dataset.Columns[idx].Type, nil
}

// SetColumnTypeAt will set column type at index `colidx` to `tipe`.
func (dataset *Dataset) SetColumnTypeAt(idx, tipe int) error {
	if idx >= dataset.GetNColumn() {
		return ErrColIdxOutOfRange
	}

	dataset.Columns[idx].Type = tipe
	return nil
}

// GetColumnsName return name of all columns.
func (dataset *Dataset) GetColumnsName() (names []string) {
	for x := range dataset.Columns {
		names = append(names, dataset.Columns[x].Name)
	}

	return
}

// SetColumnsName set column name.
func (dataset *Dataset) SetColumnsName(names []string) {
	nameslen := len(names)

	if nameslen <= 0 {
		// empty names, return immediately.
		return
	}

	collen := dataset.GetNColumn()

	if collen <= 0 {
		dataset.Columns = make(Columns, nameslen)
		collen = nameslen
	}

	// find minimum length
	minlen := collen
	if nameslen < collen {
		minlen = nameslen
	}

	for x := 0; x < minlen; x++ {
		dataset.Columns[x].Name = names[x]
	}
}

// AddColumn will create and add new empty column with specific type and name
// into dataset.
func (dataset *Dataset) AddColumn(tipe int, name string, vs []string) {
	col := Column{
		Type:       tipe,
		Name:       name,
		ValueSpace: vs,
	}
	dataset.PushColumn(col)
}

// GetColumn return pointer to column object at index `idx`.  If `idx` is out of
// range return nil.
func (dataset *Dataset) GetColumn(idx int) (col *Column) {
	if idx > dataset.GetNColumn() {
		return
	}

	switch dataset.Mode {
	case DatasetModeRows:
		dataset.TransposeToColumns()
	case DatasetModeColumns:
		// do nothing
	case DatasetModeMatrix:
		// do nothing
	}

	return &dataset.Columns[idx]
}

// GetColumnByName return column based on their `name`.
func (dataset *Dataset) GetColumnByName(name string) (col *Column) {
	if dataset.Mode == DatasetModeRows {
		dataset.TransposeToColumns()
	}

	for x, col := range dataset.Columns {
		if col.Name == name {
			return &dataset.Columns[x]
		}
	}
	return
}

// GetColumns return columns in dataset, without transposing.
func (dataset *Dataset) GetColumns() *Columns {
	return &dataset.Columns
}

// SetColumns will replace current columns with new one from parameter.
func (dataset *Dataset) SetColumns(cols *Columns) {
	dataset.Columns = *cols
}

// GetRow return pointer to row at index `idx` or nil if index is out of range.
func (dataset *Dataset) GetRow(idx int) *Row {
	if idx < 0 {
		return nil
	}
	if idx >= dataset.Rows.Len() {
		return nil
	}
	return dataset.Rows[idx]
}

// GetRows return rows in dataset, without transposing.
func (dataset *Dataset) GetRows() *Rows {
	return &dataset.Rows
}

// SetRows will replace current rows with new one from parameter.
func (dataset *Dataset) SetRows(rows *Rows) {
	dataset.Rows = *rows
}

// GetData return the data, based on mode (rows, columns, or matrix).
func (dataset *Dataset) GetData() interface{} {
	switch dataset.Mode {
	case DatasetModeRows:
		return &dataset.Rows
	case DatasetModeColumns:
		return &dataset.Columns
	case DatasetModeMatrix, DatasetNoMode:
		return &Matrix{
			Columns: &dataset.Columns,
			Rows:    &dataset.Rows,
		}
	}

	return nil
}

// GetDataAsRows return data in rows mode.
func (dataset *Dataset) GetDataAsRows() *Rows {
	if dataset.Mode == DatasetModeColumns {
		dataset.TransposeToRows()
	}
	return &dataset.Rows
}

// GetDataAsColumns return data in columns mode.
func (dataset *Dataset) GetDataAsColumns() (columns *Columns) {
	if dataset.Mode == DatasetModeRows {
		dataset.TransposeToColumns()
	}
	return &dataset.Columns
}

// TransposeToColumns move all data from rows (horizontal) to columns
// (vertical) mode.
func (dataset *Dataset) TransposeToColumns() {
	if dataset.GetNRow() <= 0 {
		// nothing to transpose
		return
	}

	ncol := dataset.GetNColumn()
	if ncol <= 0 {
		// if no columns defined, initialize it using record type
		// in the first row.
		types := dataset.GetRow(0).Types()
		dataset.SetColumnsType(types)
		ncol = len(types)
	}

	orgmode := dataset.GetMode()

	switch orgmode {
	case DatasetModeRows:
		// do nothing.
	case DatasetModeColumns, DatasetModeMatrix, DatasetNoMode:
		// check if column records contain data.
		nrow := dataset.Columns[0].Len()
		if nrow > 0 {
			// return if column record is not empty, its already
			// transposed
			return
		}
	}

	// use the least length
	minlen := len(*dataset.GetRow(0))

	if minlen > ncol {
		minlen = ncol
	}

	switch orgmode {
	case DatasetModeRows, DatasetNoMode:
		dataset.SetMode(DatasetModeColumns)
	}

	for _, row := range dataset.Rows {
		for y := 0; y < minlen; y++ {
			dataset.Columns[y].PushBack((*row)[y])
		}
	}

	// reset the rows data only if original mode is rows
	// this to prevent empty data when mode is matrix.
	if orgmode == DatasetModeRows {
		dataset.Rows = nil
	}
}

// TransposeToRows will move all data from columns (vertical) to rows
// (horizontal) mode.
func (dataset *Dataset) TransposeToRows() {
	orgmode := dataset.GetMode()

	if orgmode == DatasetModeRows {
		// already transposed
		return
	}

	if orgmode == DatasetModeColumns {
		// only set mode if transposing from columns to rows
		dataset.SetMode(DatasetModeRows)
	}

	// Get the max length of columns.
	rowlen := math.MinInt32
	flen := len(dataset.Columns)

	for f := 0; f < flen; f++ {
		l := dataset.Columns[f].Len()

		if l > rowlen {
			rowlen = l
		}
	}

	dataset.Rows = make(Rows, 0)

	// Transpose record from column to row.
	for r := 0; r < rowlen; r++ {
		row := make(Row, flen)

		for f := 0; f < flen; f++ {
			if dataset.Columns[f].Len() > r {
				row[f] = dataset.Columns[f].Records[r]
			} else {
				row[f] = NewRecord()
			}
		}

		dataset.Rows = append(dataset.Rows, &row)
	}

	// Only reset the columns if original dataset mode is "columns".
	// This to prevent empty data when mode is matrix.
	if orgmode == DatasetModeColumns {
		dataset.Columns.Reset()
	}
}

// PushRow save the data, which is already in row object, to Rows.
func (dataset *Dataset) PushRow(row *Row) {
	switch dataset.GetMode() {
	case DatasetModeRows:
		dataset.Rows = append(dataset.Rows, row)
	case DatasetModeColumns:
		dataset.PushRowToColumns(row)
	case DatasetModeMatrix, DatasetNoMode:
		dataset.Rows = append(dataset.Rows, row)
		dataset.PushRowToColumns(row)
	}
}

// PushRowToColumns push each data in Row to Columns.
func (dataset *Dataset) PushRowToColumns(row *Row) {
	rowlen := row.Len()
	if rowlen <= 0 {
		// return immediately if no data in row.
		return
	}

	// check if columns is initialize.
	collen := len(dataset.Columns)
	if collen <= 0 {
		dataset.Columns = make(Columns, rowlen)
		collen = rowlen
	}

	// pick the minimum length.
	min := rowlen
	if collen < rowlen {
		min = collen
	}

	for x := 0; x < min; x++ {
		dataset.Columns[x].PushBack((*row)[x])
	}
}

// FillRowsWithColumn given a column, fill the dataset with row where the record
// only set at index `colIdx`.
//
// Example, content of dataset was,
//
// index:	0 1 2
//
//	A B C
//	X     (step 1) nrow = 2
//
// If we filled column at index 2 with [Y Z], the dataset will become:
//
// index:	0 1 2
//
//	A B C
//	X   Y (step 2) fill the empty row
//	    Z (step 3) create dummy row which contain the rest of column data.
func (dataset *Dataset) FillRowsWithColumn(colIdx int, col Column) {
	if dataset.GetMode() != DatasetModeRows {
		// Only work if dataset mode is ROWS
		return
	}

	nrow := dataset.GetNRow()
	emptyAt := nrow

	// (step 1) Find the row with empty records
	for x, row := range dataset.Rows {
		if row.IsNilAt(colIdx) {
			emptyAt = x
			break
		}
	}

	// (step 2) Fill the empty rows using column records.
	y := 0
	for x := emptyAt; x < nrow; x++ {
		dataset.Rows[x].SetValueAt(colIdx, col.Records[y])
		y++
	}

	// (step 3) Continue filling the column but using dummy row which
	// contain only record at index `colIdx`.
	ncol := dataset.GetNColumn()
	nrow = col.Len()
	for ; y < nrow; y++ {
		row := make(Row, ncol)

		for z := 0; z < ncol; z++ {
			if z == colIdx {
				row[colIdx] = col.Records[y]
			} else {
				row[z] = NewRecord()
			}
		}

		dataset.PushRow(&row)
	}
}

// PushColumn will append new column to the end of slice if no existing column
// with the same name. If it exist, the records will be merged.
func (dataset *Dataset) PushColumn(col Column) {
	exist := false
	colIdx := 0
	for x, c := range dataset.Columns {
		if c.Name == col.Name {
			exist = true
			colIdx = x
			break
		}
	}

	switch dataset.GetMode() {
	case DatasetModeRows:
		if exist {
			dataset.FillRowsWithColumn(colIdx, col)
		} else {
			// append new column
			dataset.Columns = append(dataset.Columns, col)
			dataset.PushColumnToRows(col)
			// Remove records in column
			dataset.Columns[dataset.GetNColumn()-1].Reset()
		}
	case DatasetModeColumns:
		if exist {
			dataset.Columns[colIdx].PushRecords(col.Records)
		} else {
			dataset.Columns = append(dataset.Columns, col)
		}
	case DatasetModeMatrix, DatasetNoMode:
		if exist {
			dataset.Columns[colIdx].PushRecords(col.Records)
		} else {
			dataset.Columns = append(dataset.Columns, col)
			dataset.PushColumnToRows(col)
		}
	}
}

// PushColumnToRows add each record in column to each rows, from top to bottom.
func (dataset *Dataset) PushColumnToRows(col Column) {
	colsize := col.Len()
	if colsize <= 0 {
		// Do nothing if column is empty.
		return
	}

	nrow := dataset.GetNRow()
	if nrow <= 0 {
		// If no existing rows in dataset, initialize the rows slice.
		dataset.Rows = make(Rows, colsize)

		for nrow = 0; nrow < colsize; nrow++ {
			row := make(Row, 0)
			dataset.Rows[nrow] = &row
		}
	}

	// Pick the minimum length between column or current row length.
	minrow := nrow

	if colsize < nrow {
		minrow = colsize
	}

	// Push each record in column to each rows
	var row *Row
	var rec *Record

	for x := 0; x < minrow; x++ {
		row = dataset.Rows[x]
		rec = col.Records[x]

		row.PushBack(rec)
	}
}

// MergeColumns append columns from other dataset into current dataset.
func (dataset *Dataset) MergeColumns(other DatasetInterface) {
	othermode := other.GetMode()
	if othermode == DatasetModeRows {
		other.TransposeToColumns()
	}

	cols := other.GetDataAsColumns()
	for _, col := range *cols {
		dataset.PushColumn(col)
	}

	if othermode == DatasetModeRows {
		other.TransposeToRows()
	}
}

// MergeRows append rows from other dataset into current dataset.
func (dataset *Dataset) MergeRows(other DatasetInterface) {
	rows := other.GetDataAsRows()
	for _, row := range *rows {
		dataset.PushRow(row)
	}
}

// DeleteRow will detach row at index `i` from dataset and return it.
func (dataset *Dataset) DeleteRow(i int) (row *Row) {
	if i < 0 {
		return
	}
	if i >= dataset.Rows.Len() {
		return
	}

	orgmode := dataset.GetMode()
	if orgmode == DatasetModeColumns {
		dataset.TransposeToRows()
	}

	row = dataset.Rows.Del(i)

	if orgmode == DatasetModeColumns {
		dataset.TransposeToColumns()
	}

	if orgmode != DatasetModeRows {
		// Delete record in each columns as the same index as deleted
		// row.
		for x := range dataset.Columns {
			dataset.Columns[x].DeleteRecordAt(i)
		}
	}

	return row
}
