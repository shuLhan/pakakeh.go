// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestToFloat64(t *testing.T) {
	in := []string{"0", "1.1", "e", "3"}
	exp := []float64{0, 1.1, 0, 3}

	got := ToFloat64(in)

	test.Assert(t, "", exp, got)
}

func TestToInt64(t *testing.T) {
	in := []string{"0", "1", "e", "3.3"}
	exp := []int64{0, 1, 0, 3}

	got := ToInt64(in)

	test.Assert(t, "", exp, got)
}

func TestToStrings(t *testing.T) {
	is := make([]interface{}, 0)
	i64 := []int64{0, 1, 2, 3}
	exp := []string{"0", "1", "2", "3"}

	for _, v := range i64 {
		is = append(is, v)
	}

	got := ToStrings(is)

	test.Assert(t, "", exp, got)
}
