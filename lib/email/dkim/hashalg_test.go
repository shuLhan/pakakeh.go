// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestUnpackHashAlg(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		exp: "",
	}, {
		in: "sha512",
	}, {
		in:  "sha512:sha256",
		exp: "sha256",
	}, {
		in:  "sha512:sha256:sha1",
		exp: "sha256:sha1",
	}}

	for _, c := range cases {
		t.Log(c.in)

		algs := unpackHashAlgs([]byte(c.in))
		got := packHashAlgs(algs)

		test.Assert(t, "unpackHashAlgs", c.exp, string(got))
	}
}
