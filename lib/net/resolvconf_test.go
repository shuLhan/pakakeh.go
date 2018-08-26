// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewResolvConf(t *testing.T) {
	cases := []struct {
		path           string
		expSearchList  []string
		expNameServers []string
	}{{
		path:           "testdata/resolv.conf",
		expSearchList:  []string{"a", "b", "c", "d", "e", "f"},
		expNameServers: []string{"127.0.0.1", "1.1.1.1", "2.2.2.2"},
	}}

	for _, c := range cases {
		rc, err := NewResolvConf(c.path)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "SearchList", c.expSearchList, rc.SearchList, true)
		test.Assert(t, "NameServers", c.expNameServers, rc.NameServers, true)
	}
}
