// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package numbers

// Int64CreateSeq will create and return sequence of integer from `min` to
// `max`.
//
// E.g. if min is 0 and max is 5 then it will return `[0 1 2 3 4 5]`.
func Int64CreateSeq(min, max int64) (seq []int64) {
	for ; min <= max; min++ {
		seq = append(seq, min)
	}
	return
}
