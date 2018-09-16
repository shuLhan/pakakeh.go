// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"fmt"
	"math/rand"
	"time"
)

//
// Rows represent slice of Row.
//
type Rows []*Row

//
// Len return number of row.
//
func (rows *Rows) Len() int {
	return len(*rows)
}

//
// PushBack append record r to the end of rows.
//
func (rows *Rows) PushBack(r *Row) {
	if r != nil {
		(*rows) = append((*rows), r)
	}
}

//
// PopFront remove the head, return the record value.
//
func (rows *Rows) PopFront() (row *Row) {
	l := len(*rows)
	if l > 0 {
		row = (*rows)[0]
		(*rows) = (*rows)[1:]
	}
	return
}

//
// PopFrontAsRows remove the head and return ex-head as new rows.
//
func (rows *Rows) PopFrontAsRows() (newRows Rows) {
	row := rows.PopFront()
	if nil == row {
		return
	}
	newRows.PushBack(row)
	return
}

//
// Del will detach row at index `i` from slice and return it.
//
func (rows *Rows) Del(i int) (row *Row) {
	if i < 0 {
		return
	}
	if i >= rows.Len() {
		return
	}

	row = (*rows)[i]

	last := len(*rows) - 1
	copy((*rows)[i:], (*rows)[i+1:])
	(*rows)[last] = nil
	(*rows) = (*rows)[0:last]

	return row
}

//
// GroupByValue will group each row based on record value in index recGroupIdx
// into map of string -> *Row.
//
// WARNING: returned rows will be empty!
//
// For example, given rows with target group in column index 1,
//
// 	[1 +]
// 	[2 -]
// 	[3 -]
// 	[4 +]
//
// this function will create a map with key is string of target and value is
// pointer to sub-rows,
//
// 	+ -> [1 +]
//           [4 +]
// 	- -> [2 -]
//           [3 -]
//
//
func (rows *Rows) GroupByValue(GroupIdx int) (mapRows MapRows) {
	for {
		row := rows.PopFront()
		if nil == row {
			break
		}

		key := fmt.Sprint((*row)[GroupIdx])

		mapRows.AddRow(key, row)
	}
	return
}

//
// RandomPick row in rows until n item and return it like its has been shuffled.
// If duplicate is true, row that has been picked can be picked up again,
// otherwise it will only picked up once.
//
// This function return picked and unpicked rows and index of them.
//
func (rows *Rows) RandomPick(n int, duplicate bool) (
	picked Rows,
	unpicked Rows,
	pickedIdx []int,
	unpickedIdx []int,
) {
	rowsLen := len(*rows)

	// if duplication is not allowed, we can only select as many as rows
	// that we have.
	if n > rowsLen && !duplicate {
		n = rowsLen
	}

	rand.Seed(time.Now().UnixNano())

	for ; n >= 1; n-- {
		idx := 0
		for {
			idx = rand.Intn(len(*rows))

			if duplicate {
				// allow duplicate idx
				pickedIdx = append(pickedIdx, idx)
				break
			}

			// check if its already picked
			isPicked := false
			for _, pastIdx := range pickedIdx {
				if idx == pastIdx {
					isPicked = true
					break
				}
			}
			// get another random idx again
			if isPicked {
				continue
			}

			// bingo, we found unique idx that has not been picked.
			pickedIdx = append(pickedIdx, idx)
			break
		}

		row := (*rows)[idx]

		picked.PushBack(row)
	}

	// select unpicked rows using picked index.
	for rid := range *rows {
		// check if row index has been picked up
		isPicked := false
		for _, idx := range pickedIdx {
			if rid == idx {
				isPicked = true
				break
			}
		}
		if !isPicked {
			unpicked.PushBack((*rows)[rid])
			unpickedIdx = append(unpickedIdx, rid)
		}
	}
	return
}

//
// Contain return true and index of row, if rows has data that has the same value
// with `row`, otherwise return false and -1 as index.
//
func (rows *Rows) Contain(xrow *Row) (bool, int) {
	for x, row := range *rows {
		if xrow.IsEqual(row) {
			return true, x
		}
	}
	return false, -1
}

//
// Contains return true and indices of row, if rows has data that has the same
// value with `rows`, otherwise return false and empty indices.
//
func (rows *Rows) Contains(xrows Rows) (isin bool, indices []int) {
	// No data to compare.
	if len(xrows) <= 0 {
		return
	}

	for _, xrow := range xrows {
		isin, idx := rows.Contain(xrow)

		if isin {
			indices = append(indices, idx)
		}
	}

	// Check if indices length equal to searched rows
	if len(indices) == len(xrows) {
		return true, indices
	}

	return false, nil
}

//
// SelectWhere return all rows which column value in `colidx` is equal
// to `colval`.
//
func (rows *Rows) SelectWhere(colidx int, colval string) (selected Rows) {
	for _, row := range *rows {
		col := (*row)[colidx]
		if col.IsEqualToString(colval) {
			selected.PushBack(row)
		}
	}
	return
}

//
// String return the string representation of each row.
//
func (rows Rows) String() (s string) {
	for x := range rows {
		s += fmt.Sprint(rows[x])
	}
	return
}
