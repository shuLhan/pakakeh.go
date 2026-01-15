// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package email

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestIsValidLocal(t *testing.T) {
	cases := []struct {
		desc string
		in   []byte
		exp  bool
	}{{
		desc: "With empty local",
	}, {
		desc: "With dot at the beginning",
		in:   []byte(".local"),
	}, {
		desc: "With dot at the end",
		in:   []byte("local."),
	}, {
		desc: "With multiple dots",
		in:   []byte("loc..al"),
	}, {
		desc: "With space",
		in:   []byte("loc al"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsValidLocal(c.in)

		test.Assert(t, "IsValidLocal", c.exp, got)
	}

	specialChars.Visit(func(k byte) bool {
		local := []byte("loc")
		local = append(local, k)
		local = append(local, "al"...)

		t.Logf("With %s", local)

		got := IsValidLocal(local)

		test.Assert(t, "IsValidLocal", false, got)
		return false
	})
}
