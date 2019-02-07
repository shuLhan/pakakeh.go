// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

// nolint: lll
func TestLookupKey(t *testing.T) {
	qmethod := QueryMethod{}

	cases := []struct {
		desc     string
		sdid     string
		selector string
		exp      string
		expErr   string
	}{{
		desc:   "With empty input",
		expErr: "dkim: LookupKey: empty SDID",
	}, {
		desc:     "With empty selector",
		sdid:     "amazonses.com",
		selector: "ug7nbtf4gccmlpwj322ax3p6ow6yfsug",
		exp:      "p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCKkjP6XucgQ06cVZ89Ue/sQDu4v1/AJVd6mMK4bS2YmXk5PzWw4KWtWNUZlg77hegAChx1pG85lUbJ+x4awp28VXqRi3/jZoC6W+3ELysDvVohZPMRMadc+KVtyTiTH4BL38/8ZV9zkj4ZIaaYyiLAiYX+c3+lZQEF3rKDptRcpwIDAQAB; k=rsa;",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := LookupKey(qmethod, []byte(c.sdid), []byte(c.selector))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Key", c.exp, string(got.Bytes()), true)
	}
}
