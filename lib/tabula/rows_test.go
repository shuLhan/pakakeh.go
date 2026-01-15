// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestPushBack(t *testing.T) {
	rows, e := initRows()
	if e != nil {
		t.Fatal(e)
	}

	exp := strings.Join(rowsExpect, "")
	got := fmt.Sprint(rows)

	test.Assert(t, "", exp, got)
}

func TestPopFront(t *testing.T) {
	rows, e := initRows()
	if e != nil {
		t.Fatal(e)
	}

	l := len(rows) - 1
	for i := range rows {
		row := rows.PopFront()

		exp := rowsExpect[i]
		got := fmt.Sprint(row)

		test.Assert(t, "", exp, got)

		if i < l {
			exp = strings.Join(rowsExpect[i+1:], "")
		} else {
			exp = ""
		}
		got = fmt.Sprint(rows)

		test.Assert(t, "", exp, got)
	}

	// empty rows
	row := rows.PopFront()

	exp := "<nil>"
	got := fmt.Sprint(row)

	test.Assert(t, "", exp, got)
}

func TestPopFrontRow(t *testing.T) {
	rows, e := initRows()
	if e != nil {
		t.Fatal(e)
	}

	l := len(rows) - 1
	for i := range rows {
		newRows := rows.PopFrontAsRows()

		exp := rowsExpect[i]
		got := fmt.Sprint(newRows)

		test.Assert(t, "", exp, got)

		if i < l {
			exp = strings.Join(rowsExpect[i+1:], "")
		} else {
			exp = ""
		}
		got = fmt.Sprint(rows)

		test.Assert(t, "", exp, got)
	}

	// empty rows
	row := rows.PopFrontAsRows()

	exp := ""
	got := fmt.Sprint(row)

	test.Assert(t, "", exp, got)
}

func TestGroupByValue(t *testing.T) {
	rows, e := initRows()
	if e != nil {
		t.Fatal(e)
	}

	mapRows := rows.GroupByValue(testClassIdx)

	got := fmt.Sprint(mapRows)

	test.Assert(t, "", groupByExpect, got)
}

func TestRandomPick(t *testing.T) {
	rows, e := initRows()
	if e != nil {
		t.Fatal(e)
	}

	// random pick with duplicate
	for i := 0; i < 5; i++ {
		picked, unpicked, pickedIdx, unpickedIdx := rows.RandomPick(6,
			true)

		// check if unpicked item exist in picked items.
		isin, _ := picked.Contains(unpicked)

		if isin {
			fmt.Println("Random pick with duplicate rows")
			fmt.Println("==> picked rows   :", picked)
			fmt.Println("==> picked idx    :", pickedIdx)
			fmt.Println("==> unpicked rows :", unpicked)
			fmt.Println("==> unpicked idx  :", unpickedIdx)
			t.Fatal("random pick: unpicked is false")
		}
	}

	// random pick without duplication
	for i := 0; i < 5; i++ {
		picked, unpicked, pickedIdx, unpickedIdx := rows.RandomPick(3,
			false)

		// check if unpicked item exist in picked items.
		isin, _ := picked.Contains(unpicked)

		if isin {
			fmt.Println("Random pick with no duplicate rows")
			fmt.Println("==> picked rows   :", picked)
			fmt.Println("==> picked idx    :", pickedIdx)
			fmt.Println("==> unpicked rows :", unpicked)
			fmt.Println("==> unpicked idx  :", unpickedIdx)
			t.Fatal("random pick: unpicked is false")
		}
	}
}

func TestRowsDel(t *testing.T) {
	rows, e := initRows()
	if e != nil {
		t.Fatal(e)
	}

	// Test deleting row index out of range.
	row := rows.Del(-1)
	if row != nil {
		t.Fatal("row should be nil!")
	}

	row = rows.Del(rows.Len())
	if row != nil {
		t.Fatal("row should be nil!")
	}

	// Test deleting index that is actually exist.
	row = rows.Del(0)

	exp := strings.Join(rowsExpect[1:], "")
	got := fmt.Sprint(rows)

	test.Assert(t, "", exp, got)

	got = fmt.Sprint(row)
	test.Assert(t, "", rowsExpect[0], got)
}
