// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package strings

import (
	"errors"
	"fmt"
	"strconv"
)

// ToBytes convert slice of string into slice of slice of bytes.
func ToBytes(ss []string) (sv [][]byte) {
	for x := range len(ss) {
		sv = append(sv, []byte(ss[x]))
	}
	return sv
}

// ToFloat64 convert slice of string to slice of float64. If converted
// string return error it will set the float value to 0.
func ToFloat64(ss []string) (sv []float64) {
	var v float64
	var e error

	for _, s := range ss {
		v, e = strconv.ParseFloat(s, 64)

		if nil != e {
			v = 0
		}

		sv = append(sv, v)
	}
	return
}

// ToInt64 convert slice of string to slice of int64. If converted
// string return an error it will set the integer value to 0.
func ToInt64(ss []string) (sv []int64) {
	for _, s := range ss {
		v, e := strconv.ParseInt(s, 10, 64)

		if e == nil {
			sv = append(sv, v)
			continue
		}

		// Handle error, try to convert to float64 first.
		var ev *strconv.NumError

		if errors.As(e, &ev) && errors.Is(ev.Err, strconv.ErrSyntax) {
			f, e := strconv.ParseFloat(s, 64)
			if e == nil {
				v = int64(f)
			}
		}

		sv = append(sv, v)
	}
	return
}

// ToStrings convert slice of interface to slice of string.
func ToStrings(is []any) (vs []string) {
	for x := range len(is) {
		v := fmt.Sprintf("%v", is[x])
		vs = append(vs, v)
	}
	return
}
