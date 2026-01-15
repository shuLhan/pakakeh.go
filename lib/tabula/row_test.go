// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func createRow() (row Row) {
	dataFloat64 := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	for _, v := range dataFloat64 {
		row.PushBack(NewRecordReal(v))
	}
	return
}

func TestClone(t *testing.T) {
	row := createRow()
	rowClone := row.Clone()
	rowClone2 := row.Clone()

	test.Assert(t, "", &row, rowClone)

	// changing the clone value should not change the original copy.
	(*rowClone2)[0].SetFloat(0)
	test.Assert(t, "", &row, rowClone)
}
