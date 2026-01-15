// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>

package text

import (
	"reflect"
	"testing"
)

func TestLine_MarshalJSON(t *testing.T) {
	type testCase struct {
		exp  string
		line Line
	}

	var cases = []testCase{{
		line: Line{
			N: 1,
			V: []byte("<script>a\"\\ \b\f\n\r\tz"),
		},
		exp: `{"N":1,"V":"<script>a\"\\ \b\f\n\r\tz"}`,
	}}

	var (
		c   testCase
		got []byte
		err error
	)

	for _, c = range cases {
		got, err = c.line.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(c.exp, string(got)) {
			t.Fatalf(`want %s, got %s`, c.exp, got)
		}
	}
}
