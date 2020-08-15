// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestKeyLookupKey(t *testing.T) {
	qmethod := QueryMethod{}

	cases := []struct {
		domainName string
		exp        string
		expErr     string
	}{{
		expErr: "dkim: LookupKey: empty domain name",
	}, {
		domainName: "ug7nbtf4gccmlpwj322ax3p6ow6yfsug._domainkey.amazonses.com",
		exp:        "p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCKkjP6XucgQ06cVZ89Ue/sQDu4v1/AJVd6mMK4bS2YmXk5PzWw4KWtWNUZlg77hegAChx1pG85lUbJ+x4awp28VXqRi3/jZoC6W+3ELysDvVohZPMRMadc+KVtyTiTH4BL38/8ZV9zkj4ZIaaYyiLAiYX+c3+lZQEF3rKDptRcpwIDAQAB",
	}, {
		domainName: "20150623._domainkey.wikimedia-or-id.20150623.gappssmtp.com",
		exp:        "v=DKIM1; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2UMfREvlgajdSp3jv1tJ9nLpi/mRYnGyKC3inEQ9a7zqUjLq/yXukgpXs9AEHlvBvioxlgAVCPQQsuc1xp9+KXQGgJ8jTsn5OtKm8u+YBCt6OfvpeCpvt0l9JXMMHBNYV4c0XiPE5RHX2ltI0Av20CfEy+vMecpFtVDg4rMngjLws/ro6qT63S20A4zyVs/V19WW5F2Lulgv+l+EJzz9XummIJHOlU5n5ChcWU3Rw5RVGTtNjTZnFUaNXly3fW0ahKcG5Qc3e0Rhztp57JJQTl3OmHiMR5cHsCnrl1VnBi3kaOoQBYsSuBm+KRhMIw/X9wkLY67VLdkrwlX3xxsp6wIDAQAB; k=rsa",
	}}

	for _, c := range cases {
		t.Log(c.domainName)

		got, err := LookupKey(qmethod, c.domainName)
		if err != nil {
			serr := err.Error()
			if strings.Contains(serr, "timeout") {
				continue
			}
			test.Assert(t, "error", c.expErr, serr, true)
			continue
		}
		if got == nil {
			continue
		}

		test.Assert(t, "Key", c.exp, got.Pack(), true)
	}
}

func TestKeyParseTXT(t *testing.T) {
	cases := []struct {
		txt    string
		ttl    uint32
		exp    string
		expErr string
	}{{
		txt: "",
	}, {
		txt:    "empty",
		expErr: "dkim: missing '=': ''",
	}, {
		txt:    "p=notabase64",
		expErr: "dkim: error parsing public key: asn1: structure error: length too large",
	}, {
		txt: "k=unknown",
	}, {
		txt:    "v = DKIM1; 0=invalidtag",
		expErr: "dkim: invalid tag key: '0'",
	}, {
		txt:    "v = DKIM1; v=duplicate",
		expErr: "dkim: duplicate tag: 'v'",
	}, {
		txt:    "v = DKIM 1; s=with\x10value;",
		expErr: "dkim: invalid tag value: 'with\x10value'",
	}, {
		txt: "v = DKIM 1; n=with space;",
		exp: "v=DKIM 1; n=with space",
	}}

	for _, c := range cases {
		t.Log(c.txt)

		got, err := ParseTXT([]byte(c.txt), c.ttl)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}
		if got == nil {
			continue
		}

		test.Assert(t, "ParseTXT", c.exp, got.Pack(), true)
	}
}

func TestKeyLookupDNSTXT(t *testing.T) {
	cases := []struct {
		dname  string
		exp    string
		expErr string
	}{{
		dname: "",
	}, {
		dname:  "invalid-domain",
		expErr: "dkim: LookupKey: DNS response status: 3",
	}, {
		dname:  "www.amazon.com",
		expErr: "dkim: LookupKey: no TXT record on 'www.amazon.com'",
	}, {
		dname:  "www.google.com",
		expErr: "dkim: LookupKey: empty answer on 'www.google.com'",
	}}

	for _, c := range cases {
		t.Log(c.dname)

		got, err := lookupDNSTXT(c.dname)
		if err != nil {
			serr := err.Error()
			if strings.Contains(serr, "timeout") {
				continue
			}
			test.Assert(t, "error", c.expErr, serr, true)
			continue
		}
		if got == nil {
			continue
		}

		test.Assert(t, "ParseTXT", c.exp, got.Pack(), true)
	}
}

func TestKeyPack(t *testing.T) {
	cases := []struct {
		key *Key
		exp string
	}{{
		exp: "",
	}, {
		key: &Key{
			Public:  []byte("test"),
			Version: []byte("DKIM1"),
			HashAlgs: []HashAlg{
				HashAlgSHA256,
			},
			Notes: []byte("notes"),
			Services: [][]byte{
				[]byte("email"),
			},
			Flags: []KeyFlag{KeyFlagTesting},
		},
		exp: "v=DKIM1; p=test; h=sha256; n=notes; s=email; t=y",
	}}

	for _, c := range cases {
		got := c.key.Pack()

		test.Assert(t, "Key.Pack", c.exp, got, true)
	}
}

func TestKeySet(t *testing.T) {
	cases := []struct {
		in     *tag
		exp    string
		expErr string
	}{{
		in: nil,
	}, {
		in: &tag{
			key:   tagDNSPublicKey,
			value: []byte("invalidbase64"),
		},
		expErr: "dkim: error decode public key: illegal base64 data at input byte 12",
	}, {
		in: &tag{
			key:   tagDNSHashAlgs,
			value: []byte("sha1"),
		},
		exp: "h=sha1",
	}, {
		in: &tag{
			key:   tagDNSNotes,
			value: []byte("notes"),
		},
		exp: "n=notes",
	}, {
		in: &tag{
			key:   tagDNSServices,
			value: []byte("services:*"),
		},
		exp: "s=services:*",
	}, {
		in: &tag{
			key:   tagDNSFlags,
			value: []byte("y:s:D:s"),
		},
		exp: "t=y:s",
	}, {
		in: &tag{
			key:   tagDNSFlags,
			value: []byte("yes"),
		},
		exp: "",
	}}

	for _, c := range cases {
		if c.in != nil {
			t.Logf("%s=%s", tagKeys[c.in.key], c.in.value)
		}

		key := &Key{}
		err := key.set(c.in)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		got := key.Pack()

		test.Assert(t, "Key.set", c.exp, got, true)
	}
}
