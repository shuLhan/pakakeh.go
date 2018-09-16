// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

//
// Records define slice of pointer to Record.
//
type Records []*Record

//
// Len will return the length of records.
//
func (recs *Records) Len() int {
	return len(*recs)
}

//
// SortByIndex will sort the records using slice of index `sortedIDx` and
// return it.
//
func (recs *Records) SortByIndex(sortedIdx []int) *Records {
	sorted := make(Records, len(*recs))

	for x, v := range sortedIdx {
		sorted[x] = (*recs)[v]
	}
	return &sorted
}

//
// CountWhere return number of record where its value is equal to `v` type and
// value.
//
func (recs *Records) CountWhere(v interface{}) (c int) {
	for _, r := range *recs {
		if r.IsEqualToInterface(v) {
			c++
		}
	}
	return
}

//
// CountsWhere will return count of each value in slice `sv`.
//
func (recs *Records) CountsWhere(vs []interface{}) (counts []int) {
	for _, v := range vs {
		c := recs.CountWhere(v)
		counts = append(counts, c)
	}
	return
}
