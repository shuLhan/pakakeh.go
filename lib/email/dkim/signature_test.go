// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParse(t *testing.T) {
	cases := []struct {
		desc         string
		in           string
		expErr       string
		expRelaxed   string
		expVerifyErr string
	}{{
		desc: "RFC 6376 page 25",
		in: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1118006938;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			"\th=from:to:subject:date;\r\n" +
			"\tbh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			"\tb=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR;\r\n" +
			"\tt=1117574938; x=1118006938; c=simple;\r\n" +
			"\tz=From:foo@eng.example.net|\r\n" +
			"\t To:joe@example.com|\r\n" +
			"\t Subject:demo=20run|\r\n" +
			"\t Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			"\ti=@eng.example.net; q=dns/txt;\r\n",
		expVerifyErr: "dkim: signature is expired at '2005-06-05 21:28:58 +0000 UTC'",
	}, {
		desc: "RFC 6376 section A.2",
		in: "v=1; a=rsa-sha256; s=brisbane; d=example.com;\r\n" +
			"\tc=simple/simple; q=dns/txt; i=joe@football.example.com;\r\n" +
			"\th=Received : From : To : Subject : Date : Message-ID;\r\n" +
			"\tbh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;\r\n" +
			"\tb=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB\r\n" +
			"\t4nujc7YopdG5dWLSdNg6xNAZpOPr+kHxt1IrE+NahM6L/LbvaHut\r\n" +
			"\tKVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrIx0orEtZV\r\n" +
			"\t4bmp/YzhwvcubU4=;\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=example.com; s=brisbane;\r\n" +
			"\th=Received:From:To:Subject:Date:Message-ID;\r\n" +
			"\tbh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;\r\n" +
			"\tb=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB" +
			"4nujc7YopdG5dWLSdNg6xNAZpOPr+kHxt1IrE+NahM6L/LbvaHut" +
			"KVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrIx0orEtZV" +
			"4bmp/YzhwvcubU4=;\r\n" +
			"\tc=simple/simple;\r\n" +
			"\ti=joe@football.example.com; q=dns/txt;\r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig, err := Parse([]byte(c.in))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Signature", c.expRelaxed, string(sig.Relaxed()), true)

		err = sig.Verify()
		if err != nil {
			test.Assert(t, "Verify: error", c.expVerifyErr, err.Error(), true)
		}
	}
}
