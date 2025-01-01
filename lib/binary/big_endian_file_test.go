// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

type testBigEndian struct {
	data any
	tag  string
}

func TestOpenBigEndianFile(t *testing.T) {
	listCase := []struct {
		name     string
		expError string
		expVal   string
	}{{
		name:     `notexist`,
		expError: `OpenBigEndianFile: open notexist: no such file or directory`,
	}, {
		name:   `testdata/BigEndianFile/open`,
		expVal: "Test OpenBigEndianFile\n",
	}}

	for _, tcase := range listCase {
		bef, err := OpenBigEndianFile(tcase.name)
		if err != nil {
			test.Assert(t, tcase.name+` error`,
				tcase.expError, err.Error())
			continue
		}

		test.Assert(t, `name`, tcase.name, bef.name)
		test.Assert(t, `val`, tcase.expVal, string(bef.val))

		err = bef.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}
