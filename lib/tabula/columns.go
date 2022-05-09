// Copyright 2017m Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	libbytes "github.com/shuLhan/share/lib/bytes"
	libnumbers "github.com/shuLhan/share/lib/numbers"
)

// Columns represent slice of Column.
type Columns []Column

// Len return length of columns.
func (cols *Columns) Len() int {
	return len(*cols)
}

// Reset each data and attribute in all columns.
func (cols *Columns) Reset() {
	for x := range *cols {
		(*cols)[x].Reset()
	}
}

// SetTypes of each column. The length of type must be equal with the number of
// column, otherwise it will used the minimum length between types or columns.
func (cols *Columns) SetTypes(types []int) {
	typeslen := len(types)
	colslen := len(*cols)
	minlen := typeslen

	if colslen < minlen {
		minlen = colslen
	}

	for x := 0; x < minlen; x++ {
		(*cols)[x].Type = types[x]
	}
}

// RandomPick column in columns until n item and return it like its has been
// shuffled.  If duplicate is true, column that has been picked can be picked up
// again, otherwise it will only picked up once.
//
// This function return picked and unpicked column and index of them.
func (cols *Columns) RandomPick(n int, dup bool, excludeIdx []int) (
	picked Columns,
	unpicked Columns,
	pickedIdx []int,
	unpickedIdx []int,
) {
	excLen := len(excludeIdx)
	colsLen := len(*cols)
	allowedLen := colsLen - excLen

	// if duplication is not allowed, limit the number of selected
	// column.
	if n > allowedLen && !dup {
		n = allowedLen
	}

	for ; n >= 1; n-- {
		idx := libnumbers.IntPickRandPositive(colsLen, dup, pickedIdx,
			excludeIdx)

		pickedIdx = append(pickedIdx, idx)
		picked = append(picked, (*cols)[idx])
	}

	// select unpicked columns using picked index.
	for cid := range *cols {
		// check if column index has been picked up
		isPicked := false
		for _, idx := range pickedIdx {
			if cid == idx {
				isPicked = true
				break
			}
		}
		if !isPicked {
			unpicked = append(unpicked, (*cols)[cid])
			unpickedIdx = append(unpickedIdx, cid)
		}
	}

	return picked, unpicked, pickedIdx, unpickedIdx
}

// GetMinMaxLength given a slice of column, find the minimum and maximum column
// length among them.
func (cols *Columns) GetMinMaxLength() (min, max int) {
	for _, col := range *cols {
		collen := col.Len()
		if collen < min {
			min = collen
		} else if collen > max {
			max = collen
		}
	}
	return
}

// Join all column records value at index `row` using separator `sep` and make
// sure if there is a separator in value it will be escaped with `esc`.
//
// Given slice of columns, where row is 1 and sep is `,` and escape is `\`
//
//	  0 1 2
//	0 A B C
//	1 D , F <- row
//	2 G H I
//
// this function will return "D,\,,F" in bytes.
func (cols *Columns) Join(row int, sep, esc []byte) (v []byte) {
	for y, col := range *cols {
		if y > 0 {
			v = append(v, sep...)
		}

		rec := col.Records[row]
		recV := rec.Bytes()

		if rec.Type() == TString {
			recV, _ = libbytes.EncloseToken(recV, sep, esc, nil)
		}

		v = append(v, recV...)
	}
	return
}
