// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

package paseto

import (
	"testing"

	"kilabit.info/pakakeh.go/lib/test"
)

func TestPreAuthEncode(t *testing.T) {
	cases := []struct {
		pieces [][]byte
		exp    []byte
	}{{
		exp: []byte("\x00\x00\x00\x00\x00\x00\x00\x00"),
	}, {
		pieces: [][]byte{[]byte(``)},
		exp:    []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
	}, {
		pieces: [][]byte{{}, {}},
		exp:    []byte("\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
	}, {
		pieces: [][]byte{[]byte(`test`)},
		exp:    []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x04\x00\x00\x00\x00\x00\x00\x00test"),
	}, {
		pieces: [][]byte{[]byte(`Paragon`)},
		exp:    []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x07\x00\x00\x00\x00\x00\x00\x00\x50\x61\x72\x61\x67\x6f\x6e"),
	}}

	for _, c := range cases {
		got, err := PreAuthEncode(c.pieces...)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, `PreAuthEncode`, c.exp, got)
	}
}
