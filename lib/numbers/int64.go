// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

//
// Int64CreateSeq will create and return sequence of integer from `min` to
// `max`.
//
// E.g. if min is 0 and max is 5 then it will return `[0 1 2 3 4 5]`.
//
func Int64CreateSeq(min, max int64) (seq []int64) {
	for ; min <= max; min++ {
		seq = append(seq, min)
	}
	return
}
