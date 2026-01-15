// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

// TestRecord simply check how the stringer work.
func TestRecord(t *testing.T) {
	expec := []string{"test", "1", "2"}
	expType := []int{TString, TInteger, TInteger}

	row := make(Row, 0)

	for i := range expec {
		r, e := NewRecordBy(expec[i], expType[i])
		if nil != e {
			t.Error(e)
		}

		row = append(row, r)
	}

	exp := fmt.Sprint(expec)
	got := fmt.Sprint(row)
	test.Assert(t, "", exp, got)
}
