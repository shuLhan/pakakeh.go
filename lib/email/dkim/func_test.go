// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestDecodeQP(t *testing.T) {
	cases := []struct {
		in  []byte
		exp []byte
	}{{
		in: []byte{},
	}, {
		in:  []byte("="),
		exp: []byte("="),
	}, {
		in:  []byte("=2"),
		exp: []byte("=2"),
	}, {
		in:  []byte("=20"),
		exp: []byte(" "),
	}, {
		in:  []byte("A=20B"),
		exp: []byte("A B"),
	}, {
		in:  []byte("A\r\n=20B"),
		exp: []byte("A B"),
	}, {
		in:  []byte("A\r\n=2xB"),
		exp: []byte("A=2xB"),
	}}

	for _, c := range cases {
		t.Log(c.in)

		got := DecodeQP(c.in)

		test.Assert(t, "DecodeQP", c.exp, got)
	}
}

func TestCanonicalize(t *testing.T) {
	cases := []struct {
		desc string
		in   string
		exp  string
	}{{
		desc: "Without 'b=' tag",
		in: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1118006938;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n",
	}, {
		desc: "RFC 6376 page 25",
		in: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1118006938;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR\r\n",
		exp: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1118006938;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			" b=",
	}, {
		desc: "RFC 6376 page 25",
		in: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1118006938;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n",
		exp: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1118006938;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" b=;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;",
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
		exp: "v=1; a=rsa-sha256; s=brisbane; d=example.com;\r\n" +
			"\tc=simple/simple; q=dns/txt; i=joe@football.example.com;\r\n" +
			"\th=Received : From : To : Subject : Date : Message-ID;\r\n" +
			"\tbh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;\r\n" +
			"\tb=;",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := Canonicalize([]byte(c.in))

		test.Assert(t, "Canonicalize", c.exp, string(got))
	}
}
