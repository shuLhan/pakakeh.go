// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestInt64CreateSeq(t *testing.T) {
	exp := []int64{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}
	got := Int64CreateSeq(-5, 5)

	test.Assert(t, "", exp, got)
}
