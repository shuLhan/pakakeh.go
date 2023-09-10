// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"fmt"
	"strconv"

	"github.com/shuLhan/share/lib/ints"
	libstrings "github.com/shuLhan/share/lib/strings"
)

// Claset define a dataset with class attribute.
type Claset struct {
	// major contain the name of majority class in dataset.
	major string

	// minor contain the name of minority class in dataset.
	minor string

	// vs contain a copy of value space.
	vs []string

	// counts number of value space in current set.
	counts []int

	// Dataset embedded, for implementing the dataset interface.
	Dataset

	// ClassIndex contain index for target classification in columns.
	ClassIndex int `json:"ClassIndex"`
}

// NewClaset create and return new Claset object.
func NewClaset(mode int, types []int, names []string) (claset *Claset) {
	claset = &Claset{
		ClassIndex: -1,
	}

	claset.Init(mode, types, names)

	return
}

// Clone return a copy of current claset object.
func (claset *Claset) Clone() interface{} {
	clone := Claset{
		ClassIndex: claset.GetClassIndex(),
		major:      claset.MajorityClass(),
		minor:      claset.MinorityClass(),
	}
	clone.SetDataset(claset.GetDataset().Clone().(DatasetInterface))
	return &clone
}

// GetDataset return the dataset.
func (claset *Claset) GetDataset() DatasetInterface {
	return &claset.Dataset
}

// GetClassType return type of class in dataset.
func (claset *Claset) GetClassType() int {
	if claset.Columns.Len() <= 0 {
		return TString
	}
	return claset.Columns[claset.ClassIndex].Type
}

// GetClassValueSpace return the class value space.
func (claset *Claset) GetClassValueSpace() []string {
	if claset.Columns.Len() <= 0 {
		return nil
	}
	return claset.Columns[claset.ClassIndex].ValueSpace
}

// GetClassColumn return dataset class values in column.
func (claset *Claset) GetClassColumn() *Column {
	if claset.Mode == DatasetModeRows {
		claset.TransposeToColumns()
	}
	if claset.Columns.Len() <= 0 {
		return nil
	}
	return &claset.Columns[claset.ClassIndex]
}

// GetClassRecords return class values as records.
func (claset *Claset) GetClassRecords() *Records {
	if claset.Mode == DatasetModeRows {
		claset.TransposeToColumns()
	}
	if claset.Columns.Len() <= 0 {
		return nil
	}
	return &claset.Columns[claset.ClassIndex].Records
}

// GetClassAsStrings return all class values as slice of string.
func (claset *Claset) GetClassAsStrings() []string {
	if claset.Mode == DatasetModeRows {
		claset.TransposeToColumns()
	}
	if claset.Columns.Len() <= 0 {
		return nil
	}
	return claset.Columns[claset.ClassIndex].ToStringSlice()
}

// GetClassAsReals return class record value as slice of float64.
func (claset *Claset) GetClassAsReals() []float64 {
	if claset.Mode == DatasetModeRows {
		claset.TransposeToColumns()
	}
	if claset.Columns.Len() <= 0 {
		return nil
	}
	return claset.Columns[claset.ClassIndex].ToFloatSlice()
}

// GetClassAsInteger return class record value as slice of int64.
func (claset *Claset) GetClassAsInteger() []int64 {
	if claset.Mode == DatasetModeRows {
		claset.TransposeToColumns()
	}
	if claset.Columns.Len() <= 0 {
		return nil
	}
	return claset.Columns[claset.ClassIndex].ToIntegers()
}

// GetClassIndex return index of class attribute in dataset.
func (claset *Claset) GetClassIndex() int {
	return claset.ClassIndex
}

// MajorityClass return the majority class of data.
func (claset *Claset) MajorityClass() string {
	return claset.major
}

// MinorityClass return the minority class in dataset.
func (claset *Claset) MinorityClass() string {
	return claset.minor
}

// Counts return the number of each class in value-space.
func (claset *Claset) Counts() []int {
	if len(claset.counts) == 0 {
		claset.CountValueSpaces()
	}
	return claset.counts
}

// SetDataset in class set.
func (claset *Claset) SetDataset(dataset DatasetInterface) {
	claset.Dataset = *(dataset.(*Dataset))
}

// SetClassIndex will set the class index to `v`.
func (claset *Claset) SetClassIndex(v int) {
	claset.ClassIndex = v
}

// SetMajorityClass will set the majority class to `v`.
func (claset *Claset) SetMajorityClass(v string) {
	claset.major = v
}

// SetMinorityClass will set the minority class to `v`.
func (claset *Claset) SetMinorityClass(v string) {
	claset.minor = v
}

// CountValueSpaces will count number of value space in current dataset.
func (claset *Claset) CountValueSpaces() {
	classv := claset.GetClassAsStrings()
	claset.vs = claset.GetClassValueSpace()

	claset.counts = libstrings.CountTokens(classv, claset.vs, false)
}

// RecountMajorMinor recount major and minor class in claset.
func (claset *Claset) RecountMajorMinor() {
	claset.CountValueSpaces()

	_, maxIdx, maxok := ints.Max(claset.counts)
	_, minIdx, minok := ints.Min(claset.counts)

	if maxok {
		claset.major = claset.vs[maxIdx]
	}
	if minok {
		claset.minor = claset.vs[minIdx]
	}
}

// IsInSingleClass check whether all target class contain only single value.
// Return true and name of target if all rows is in the same class,
// false and empty string otherwise.
func (claset *Claset) IsInSingleClass() (single bool, class string) {
	classv := claset.GetClassAsStrings()

	for i, t := range classv {
		if i == 0 {
			single = true
			class = t
			continue
		}
		if t != class {
			return false, ""
		}
	}
	return
}

// GetMinorityRows return rows where their class is minority in dataset, or nil
// if dataset is empty.
func (claset *Claset) GetMinorityRows() *Rows {
	if claset.Len() == 0 {
		return nil
	}
	if claset.vs == nil {
		claset.RecountMajorMinor()
	}

	minRows := claset.GetRows().SelectWhere(claset.ClassIndex,
		claset.minor)

	return &minRows
}

// String, yes it will pretty print the meta-data in JSON format.
func (claset *Claset) String() (s string) {
	if claset.vs == nil {
		claset.RecountMajorMinor()
	}

	s = fmt.Sprintf("'claset':{'rows': %d, 'columns': %d, ", claset.Len(),
		claset.GetNColumn())

	s += "'vs':{"
	for x, v := range claset.vs {
		if x > 0 {
			s += ", "
		}
		s += "'" + v + "':" + strconv.Itoa(claset.counts[x])
	}
	s += "}"

	s += ", 'major': '" + claset.major + "'"
	s += ", 'minor': '" + claset.minor + "'"
	s += "}"

	return
}
