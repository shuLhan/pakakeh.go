// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package strings

import (
	"strings"
)

// Row is simplified name for slice of slice of string.
type Row [][]string

// IsEqual compare two row without regard to their order.
//
// Return true if both contain the same list, false otherwise.
func (row Row) IsEqual(b Row) bool {
	rowlen := len(row)

	if rowlen != len(b) {
		return false
	}

	check := make([]bool, rowlen)

	for x, row := range row {
		for _, rstrings := range b {
			if IsEqual(row, rstrings) {
				check[x] = true
				break
			}
		}
	}

	for _, v := range check {
		if !v {
			return false
		}
	}
	return true
}

// Join list of slice of string using `lsep` as separator between row items
// and `ssep` for element in each item.
func (row Row) Join(lsep string, ssep string) (s string) {
	for x := range len(row) {
		if x > 0 {
			s += lsep
		}
		s += strings.Join(row[x], ssep)
	}
	return
}
