// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var dataFloat64 = []float64{0.1, 0.2, 0.3, 0.4, 0.5}

func createRow() (row Row) {
	for _, v := range dataFloat64 {
		row.PushBack(NewRecordReal(v))
	}
	return
}

func TestClone(t *testing.T) {
	row := createRow()
	rowClone := row.Clone()
	rowClone2 := row.Clone()

	test.Assert(t, "", &row, rowClone, true)

	// changing the clone value should not change the original copy.
	(*rowClone2)[0].SetFloat(0)
	test.Assert(t, "", &row, rowClone, true)
	test.Assert(t, "", &row, rowClone2, false)
}
