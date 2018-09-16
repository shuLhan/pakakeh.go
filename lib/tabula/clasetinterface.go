// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

//
// ClasetInterface is the interface for working with dataset containing class
// or target attribute. It embed dataset interface.
//
// Yes, the name is Claset with single `s` not Classset with triple `s` to
// minimize typo.
//
type ClasetInterface interface {
	DatasetInterface

	GetClassType() int
	GetClassValueSpace() []string
	GetClassColumn() *Column
	GetClassRecords() *Records
	GetClassAsStrings() []string
	GetClassAsReals() []float64
	GetClassIndex() int
	MajorityClass() string
	MinorityClass() string
	Counts() []int

	SetDataset(DatasetInterface)
	SetClassIndex(int)
	SetMajorityClass(string)
	SetMinorityClass(string)

	CountValueSpaces()
	RecountMajorMinor()
	IsInSingleClass() (bool, string)

	GetMinorityRows() *Rows
}
