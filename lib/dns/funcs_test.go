// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestGetSystemNameServers(t *testing.T) {
	cases := []struct {
		path string
		exp  []string
	}{{
		path: "testdata/resolv.conf",
		exp: []string{
			"127.0.0.1",
		},
	}}

	for _, c := range cases {
		t.Log(c.path)

		got := GetSystemNameServers(c.path)

		test.Assert(t, "NameServers", c.exp, got, true)
	}
}
