// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
