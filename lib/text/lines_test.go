// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package text

import (
	"reflect"
	"testing"
)

func TestParseLines(t *testing.T) {
	cases := []struct {
		raw []byte
		exp Lines
	}{{
		raw: []byte(`with single line`),
		exp: Lines{{
			V: []byte(`with single line`),
		}},
	}, {
		raw: []byte(`with
multiple
lines
`),
		exp: Lines{{
			V: []byte(`with`),
		}, {
			N: 1,
			V: []byte(`multiple`),
		}, {
			N: 2,
			V: []byte(`lines`),
		}},
	}}

	for _, c := range cases {
		got := ParseLines(c.raw)
		if !reflect.DeepEqual(c.exp, got) {
			t.Fatalf(`want %s, got %s`, c.exp, got)
		}
	}
}
