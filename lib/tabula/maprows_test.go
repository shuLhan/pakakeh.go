// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestAddRow(t *testing.T) {
	mapRows := MapRows{}
	rows, e := initRows()

	if e != nil {
		t.Fatal(e)
	}

	for _, row := range rows {
		key := fmt.Sprint((*row)[testClassIdx].Interface())
		mapRows.AddRow(key, row)
	}

	got := fmt.Sprint(mapRows)

	test.Assert(t, "", groupByExpect, got)
}

func TestGetMinority(t *testing.T) {
	mapRows := MapRows{}
	rows, e := initRows()

	if e != nil {
		t.Fatal(e)
	}

	for _, row := range rows {
		key := fmt.Sprint((*row)[testClassIdx].Interface())
		mapRows.AddRow(key, row)
	}

	// remove the first row in the first key, so we can make it minority.
	mapRows[0].Value.PopFront()

	_, minRows := mapRows.GetMinority()

	exp := rowsExpect[3]
	got := fmt.Sprint(minRows)

	test.Assert(t, "", exp, got)
}
