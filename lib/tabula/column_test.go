// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func initColReal(t *testing.T) (col *Column) {
	data := []string{"9.987654321", "8.8", "7.7", "6.6", "5.5", "4.4", "3.3"}
	col = NewColumn(TReal, "TREAL")

	for x := range data {
		rec, e := NewRecordBy(data[x], TReal)
		if e != nil {
			t.Fatal(e)
		}

		col.PushBack(rec)
	}

	return col
}

func TestToFloatSlice(t *testing.T) {
	col := initColReal(t)
	got := col.ToFloatSlice()
	expFloat := []float64{9.987654321, 8.8, 7.7, 6.6, 5.5, 4.4, 3.3}

	test.Assert(t, "", expFloat, got)
}

func TestToStringSlice(t *testing.T) {
	var col Column

	data := []string{"9.987654321", "8.8", "7.7", "6.6", "5.5", "4.4", "3.3"}

	for x := range data {
		rec, e := NewRecordBy(data[x], TString)
		if e != nil {
			t.Fatal(e)
		}

		col.PushBack(rec)
	}

	got := col.ToStringSlice()

	test.Assert(t, "", data, got)
}

func TestDeleteRecordAt(t *testing.T) {
	var exp []float64
	del := 2
	expFloat := []float64{9.987654321, 8.8, 7.7, 6.6, 5.5, 4.4, 3.3}

	exp = append(exp, expFloat[:del]...)
	exp = append(exp, expFloat[del+1:]...)

	col := initColReal(t)
	col.DeleteRecordAt(del)
	got := col.ToFloatSlice()

	test.Assert(t, "", exp, got)
}
