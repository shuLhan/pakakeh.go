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
		sdid     string
		selector string
		exp      string
		expErr   string
	}{{
		expErr: "dkim: LookupKey: empty SDID",
	}, {
		sdid:     "amazonses.com",
		selector: "ug7nbtf4gccmlpwj322ax3p6ow6yfsug",
		exp:      "p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCKkjP6XucgQ06cVZ89Ue/sQDu4v1/AJVd6mMK4bS2YmXk5PzWw4KWtWNUZlg77hegAChx1pG85lUbJ+x4awp28VXqRi3/jZoC6W+3ELysDvVohZPMRMadc+KVtyTiTH4BL38/8ZV9zkj4ZIaaYyiLAiYX+c3+lZQEF3rKDptRcpwIDAQAB; k=rsa;",
	}, {
		sdid:     "wikimedia-or-id.20150623.gappssmtp.com",
		selector: "20150623",
		exp:      "v=DKIM1; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2UMfREvlgajdSp3jv1tJ9nLpi/mRYnGyKC3inEQ9a7zqUjLq/yXukgpXs9AEHlvBvioxlgAVCPQQsuc1xp9+KXQGgJ8jTsn5OtKm8u+YBCt6OfvpeCpvt0l9JXMMHBNYV4c0XiPE5RHX2ltI0Av20CfEy+vMecpFtVDg4rMngjLws/ro6qT63S20A4zyVs/V19WW5F2Lulgv+l+EJzz9XummIJHOlU5n5ChcWU3Rw5RVGTtNjTZnFUaNXly3fW0ahKcG5Qc3e0Rhztp57JJQTl3OmHiMR5cHsCnrl1VnBi3kaOoQBYsSuBm+KRhMIw/X9wkLY67VLdkrwlX3xxsp6wIDAQAB; k=rsa;",
	}}

	for _, c := range cases {
		t.Logf("%s._domainkey.%s", c.selector, c.sdid)

		got, err := LookupKey(qmethod, []byte(c.sdid), []byte(c.selector))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Key", c.exp, string(got.Bytes()), true)
	}
}
