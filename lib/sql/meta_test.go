// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

package sql

import (
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestMetaBind(t *testing.T) {
	type bindCase struct {
		val  any
		name string
	}
	type testCase struct {
		desc         string
		driver       string
		expListName  string
		expListValue string
		expIndex     string
		kind         DMLKind
		bind         []bindCase
		expNholder   int
	}

	var listCase = []testCase{{
		desc:   `With the same name`,
		driver: DriverNamePostgres,
		kind:   DMLKindInsert,
		bind: []bindCase{{
			name: `f1`,
			val:  `v1`,
		}, {
			name: `f1`,
			val:  `v1.1`,
		}},
		expListName:  `[f1]`,
		expListValue: `[v1.1]`,
		expIndex:     `[1]`,
		expNholder:   1,
	}}

	var c testCase

	for _, c = range listCase {
		var meta = NewMeta(c.driver, c.kind)

		var bind bindCase

		for _, bind = range c.bind {
			meta.Bind(bind.name, bind.val)
		}

		var got = fmt.Sprintf(`%v`, meta.ListName)
		test.Assert(t, `ListName`, c.expListName, got)

		got = fmt.Sprintf(`%v`, meta.ListValue)
		test.Assert(t, `ListValue`, c.expListValue, got)

		got = fmt.Sprintf(`%v`, meta.Index)
		test.Assert(t, `Index`, c.expIndex, got)

		test.Assert(t, `nholder`, c.expNholder, meta.nholder)
	}
}
