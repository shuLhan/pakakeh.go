// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewTag(t *testing.T) {
	cases := []struct {
		key    string
		exp    *tag
		expErr string
	}{{
		key: "",
	}, {
		key:    "0tag",
		expErr: "dkim: invalid tag key: '0tag'",
	}, {
		key:    "a-b",
		expErr: "dkim: invalid tag key: 'a-b'",
	}}

	for _, c := range cases {
		t.Log(c.key)

		got, err := newTag([]byte(c.key))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}
		if got == nil {
			continue
		}

		test.Assert(t, "tag", c.exp, got, true)
	}
}
