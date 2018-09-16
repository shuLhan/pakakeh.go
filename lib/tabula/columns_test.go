// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
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

	// random pick with duplicate
	ncols := 6
	dup := true
	excludeIdx := []int{3}

	for i := 0; i < 5; i++ {
		picked, unpicked, _, _ :=
			dataset.Columns.RandomPick(ncols, dup, excludeIdx)

		// check if unpicked item exist in picked items.
		for _, un := range unpicked {
			for _, pick := range picked {
				test.Assert(t, "", un, pick, false)
			}
		}
	}

	// random pick without duplicate
	dup = false
	for i := 0; i < 5; i++ {
		picked, unpicked, _, _ :=
			dataset.Columns.RandomPick(ncols, dup, excludeIdx)

		// check if unpicked item exist in picked items.
		for _, un := range unpicked {
			for _, pick := range picked {
				test.Assert(t, "", un, pick, false)
			}
		}
	}
}
