// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestSortByIndex(t *testing.T) {
	data := make(Records, 3)
	data[0] = NewRecordInt(3)
	data[1] = NewRecordInt(2)
	data[2] = NewRecordInt(1)

	sortedIdx := []int{2, 1, 0}
	expect := []int{1, 2, 3}

	sorted := data.SortByIndex(sortedIdx)

	got := fmt.Sprint(sorted)
	exp := fmt.Sprint(&expect)

	test.Assert(t, "", exp, got)
}
