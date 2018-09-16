// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"strconv"
)

//
// Column represent slice of record. A vertical representation of data.
//
type Column struct {
	// Name of column. String identifier for the column.
	Name string
	// Type of column. All record in column have the same type.
	Type int
	// Flag additional attribute that can be set to mark some value on this
	// column
	Flag int
	// ValueSpace contain the possible value in records
	ValueSpace []string
	// Records contain column data.
	Records Records
}

//
// NewColumn return new column with type and name.
//
func NewColumn(colType int, colName string) (col *Column) {
	col = &Column{
		Type: colType,
		Name: colName,
		Flag: 0,
	}

	col.Records = make([]*Record, 0)

	return
}

//
// NewColumnString initialize column with type anda data as string.
//
func NewColumnString(data []string, colType int, colName string) (
	col *Column,
	e error,
) {
	col = NewColumn(colType, colName)

	datalen := len(data)

	if datalen <= 0 {
		return
	}

	col.Records = make([]*Record, datalen)

	for x := 0; x < datalen; x++ {
		col.Records[x] = NewRecordString(data[x])
	}

	return col, nil
}

//
// NewColumnInt create new column with record type as integer, and fill it
// with `data`.
//
func NewColumnInt(data []int64, colName string) (col *Column) {
	col = NewColumn(TInteger, colName)

	datalen := len(data)
	if datalen <= 0 {
		return
	}

	col.Records = make([]*Record, datalen)

	for x, v := range data {
		col.Records[x] = NewRecordInt(v)
	}
	return
}

//
// NewColumnReal create new column with record type is real.
//
func NewColumnReal(data []float64, colName string) (col *Column) {
	col = NewColumn(TReal, colName)

	datalen := len(data)

	if datalen <= 0 {
		return
	}

	col.Records = make([]*Record, datalen)

	for x := 0; x < datalen; x++ {
		rec := NewRecordReal(data[x])
		col.Records[x] = rec
	}

	return
}

//
// SetType will set the type of column to `tipe`.
//
func (col *Column) SetType(tipe int) {
	col.Type = tipe
}

//
// SetName will set the name of column to `name`.
//
func (col *Column) SetName(name string) {
	col.Name = name
}

//
// GetType return the type of column.
//
func (col *Column) GetType() int {
	return col.Type
}

//
// GetName return the column name.
//
func (col *Column) GetName() string {
	return col.Name
}

//
// SetRecords will set records in column to `recs`.
//
func (col *Column) SetRecords(recs *Records) {
	col.Records = *recs
}

//
// Interface return the column object as an interface.
//
func (col *Column) Interface() interface{} {
	return col
}

//
// Reset column data and flag.
//
func (col *Column) Reset() {
	col.Flag = 0
	col.Records = make([]*Record, 0)
}

//
// Len return number of record.
//
func (col *Column) Len() int {
	return len(col.Records)
}

//
// PushBack push record the end of column.
//
func (col *Column) PushBack(r *Record) {
	col.Records = append(col.Records, r)
}

//
// PushRecords append slice of record to the end of column's records.
//
func (col *Column) PushRecords(rs []*Record) {
	col.Records = append(col.Records, rs...)
}

//
// ToIntegers convert slice of record to slice of int64.
//
func (col *Column) ToIntegers() []int64 {
	newcol := make([]int64, col.Len())

	for x := range col.Records {
		newcol[x] = col.Records[x].Integer()
	}

	return newcol
}

//
// ToFloatSlice convert slice of record to slice of float64.
//
func (col *Column) ToFloatSlice() (newcol []float64) {
	newcol = make([]float64, col.Len())

	for i := range col.Records {
		newcol[i] = col.Records[i].Float()
	}

	return
}

//
// ToStringSlice convert slice of record to slice of string.
//
func (col *Column) ToStringSlice() (newcol []string) {
	newcol = make([]string, col.Len())

	for i := range col.Records {
		newcol[i] = col.Records[i].String()
	}

	return
}

//
// ClearValues set all value in column to empty string or zero if column type is
// numeric.
//
func (col *Column) ClearValues() {
	for _, r := range col.Records {
		r.Reset()
	}
}

//
// SetValueAt will set column value at cell `idx` with `v`, unless the index
// is out of range.
//
func (col *Column) SetValueAt(idx int, v string) {
	if idx < 0 {
		return
	}
	if col.Records.Len() <= idx {
		return
	}
	_ = col.Records[idx].SetValue(v, col.Type)
}

//
// SetValueByNumericAt will set column value at cell `idx` with numeric value
// `v`, unless the index is out of range.
//
func (col *Column) SetValueByNumericAt(idx int, v float64) {
	if idx < 0 {
		return
	}
	if col.Records.Len() <= idx {
		return
	}
	switch col.Type {
	case TString:
		col.Records[idx].SetString(strconv.FormatFloat(v, 'f', -1, 64))
	case TInteger:
		col.Records[idx].SetInteger(int64(v))
	case TReal:
		col.Records[idx].SetFloat(v)
	}
}

//
// SetValues of all column record.
//
func (col *Column) SetValues(values []string) {
	vallen := len(values)
	reclen := col.Len()

	// initialize column record if its empty.
	if reclen <= 0 {
		col.Records = make([]*Record, vallen)
		reclen = vallen
	}

	// pick the least length
	minlen := reclen
	if vallen < reclen {
		minlen = vallen
	}

	for x := 0; x < minlen; x++ {
		_ = col.Records[x].SetValue(values[x], col.Type)
	}
}

//
// DeleteRecordAt will delete record at index `i` and return it.
//
func (col *Column) DeleteRecordAt(i int) *Record {
	if i < 0 {
		return nil
	}

	clen := col.Len()
	if i >= clen {
		return nil
	}

	r := col.Records[i]

	last := clen - 1
	copy(col.Records[i:], col.Records[i+1:])
	col.Records[last] = nil
	col.Records = col.Records[0:last]

	return r
}
