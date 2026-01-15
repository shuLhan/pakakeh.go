// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
)

func TestRandomPickColumns(t *testing.T) {
	var dataset Dataset
	var e error

	dataset.Init(DatasetModeRows, testColTypes, testColNames)

	dataset.Rows, e = initRows()
	if e != nil {
		t.Fatal(e)
	}

	dataset.TransposeToColumns()

	// random pick without duplicate
	dup := false
	ncols := 6
	excludeIdx := []int{3}
	for i := 0; i < 5; i++ {
		picked, unpicked, _, _ :=
			dataset.Columns.RandomPick(ncols, dup, excludeIdx)

		// check if unpicked item exist in picked items.
		for _, un := range unpicked {
			for _, pick := range picked {
				err := reflect.DoEqual(un, pick)
				if err == nil {
					t.Fatalf("unpicked column exist in picked: %v", un)
				}
			}
		}
	}
}
