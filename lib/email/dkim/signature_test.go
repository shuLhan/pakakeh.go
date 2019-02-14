// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestSignatureParse(t *testing.T) {
	cases := []struct {
		desc           string
		in             string
		expErr         string
		expPack        string
		expSimple      string
		expValidateErr string
	}{{
		desc: "With empty input",
	}, {
		desc:   "Without end with CRLF",
		in:     "v=1;",
		expErr: "dkim: value must end with CRLF",
	}, {
		desc:   "With invalid tag",
		in:     "v=1; 0=invalidtag\r\n",
		expErr: "dkim: invalid tag key: '0'",
	}, {
		desc:   "With invalid version",
		in:     "v=2; 0=invalidtag\r\n",
		expErr: "dkim: invalid version: '2'",
	}, {
		desc:   "With invalid algorithm",
		in:     "a=sha-256\r\n",
		expErr: "dkim: unknown algorithm: 'sha-256'",
	}, {
		desc:   "With invalid SDID",
		in:     "v=1; a=rsa-sha256; d=\r\n\t  ; h=form;\r\n",
		expErr: errEmptySDID.Error(),
	}, {
		desc:   "With empty selector",
		in:     "v=1; a=rsa-sha256; d=x; s=;\r\n",
		expErr: errEmptySelector.Error(),
	}, {
		desc:   "With no From in header",
		in:     "v=1; a=rsa-sha256; d=x; h=to:subject;\r\n",
		expErr: errFromHeader.Error(),
	}, {
		desc:   "With empty header",
		in:     "v=1; a=rsa-sha256; d=x; h=; s=;\r\n",
		expErr: errEmptyHeader.Error(),
	}, {
		desc:   "With empty body hash",
		in:     "v=1; a=rsa-sha256; d=x; bh=;\r\n",
		expErr: errEmptyBodyHash.Error(),
	}, {
		desc:   "With empty signature",
		in:     "v=1; a=rsa-sha256; d=x; bh=bh; b=;\r\n",
		expErr: errEmptySignature.Error(),
	}, {
		desc:   "With invalid creation time",
		in:     "v=1; a=rsa-sha256; d=x; bh=bh; b=b; t=xxx;\r\n",
		expErr: `dkim: t=: strconv.ParseUint: parsing "xxx": invalid syntax`,
	}, {
		desc:   "With invalid expiration time",
		in:     "v=1; a=rsa-sha256; d=x; bh=bh; b=b; t=1; x=2a;\r\n",
		expErr: `dkim: x=: strconv.ParseUint: parsing "2a": invalid syntax`,
	}, {
		desc:   "With invalid canon",
		in:     "a=rsa-sha256;\r\nc=s/s/s\r\n",
		expErr: "dkim: invalid canonicalization: 's/s/s'",
	}, {
		desc:   "With invalid canon name",
		in:     "a=rsa-sha256;\r\nc=s/simple\r\n",
		expErr: "dkim: invalid canonicalization: 's'",
	}, {
		desc:   "With invalid canon name",
		in:     "a=rsa-sha256;\r\nc=simple/relax\r\n",
		expErr: "dkim: invalid canonicalization: 'relax'",
	}, {
		desc:           "With invalid query method",
		in:             "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=/;\r\n",
		expSimple:      "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=/;\r\n",
		expPack:        "v=1; a=rsa-sha256; d=x; s=s;\r\n\th=from;\r\n\tbh=bh;\r\n\tb=h;\r\n\t\r\n",
		expValidateErr: "dkim: invalid version: '2'",
	}, {
		desc:      "With invalid query type",
		in:        "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=s/;\r\n",
		expSimple: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=s/;\r\n",
		expPack:   "v=1; a=rsa-sha256; d=x; s=s;\r\n\th=from;\r\n\tbh=bh;\r\n\tb=h;\r\n\t\r\n",
	}, {
		desc:      "With invalid query option",
		in:        "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/relax;\r\n",
		expSimple: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/relax;\r\n",
		expPack:   "v=1; a=rsa-sha256; d=x; s=s;\r\n\th=from;\r\n\tbh=bh;\r\n\tb=h;\r\n\t\r\n",
	}, {
		desc:      "Without query type",
		in:        "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns;\r\n",
		expSimple: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns;\r\n",
		expPack:   "v=1; a=rsa-sha256; d=x; s=s;\r\n\th=from;\r\n\tbh=bh;\r\n\tb=h;\r\n\tq=dns/txt;\r\n",
	}, {
		desc:      "With empty query type",
		in:        "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/;\r\n",
		expSimple: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/;\r\n",
		expPack:   "v=1; a=rsa-sha256; d=x; s=s;\r\n\th=from;\r\n\tbh=bh;\r\n\tb=h;\r\n\tq=dns/txt;\r\n",
	}, {
		desc:           "With unknown tag",
		in:             "v=1;\r\n j=unknown;\r\n a=rsa-sha256; l=512\r\n",
		expPack:        "v=1; a=rsa-sha256; d=; s=;\r\n\th=;\r\n\tbh=;\r\n\tb=;\r\n\tl=512; \r\n",
		expSimple:      "v=1;\r\n j=unknown;\r\n a=rsa-sha256; l=512\r\n",
		expValidateErr: errEmptySDID.Error(),
	}, {
		desc:   "Without domain in AUID",
		in:     "v=1; a=rsa-sha256; d=x; bh=bh; b=b; i=my-auid\r\n",
		expErr: "dkim: no domain in AUID 'i=' tag: 'my-auid'",
	}, {
		desc:   "With invalid AUID",
		in:     "v=1; a=rsa-sha256; d=x; bh=bh; b=b; i=my@x.com\r\n",
		expErr: "dkim: invalid AUID: 'my@x.com'",
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
		expErr: "dkim: signature is expired at '2005-06-05 21:28:58 +0000 UTC'",
	}, {
		desc: "RFC 6376 page 25, with valid expiration time",
		in: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1577811600;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR\r\n",
		expSimple: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
			" t=1117574938; x=1577811600;\r\n" +
			" h=from:to:subject:date;\r\n" +
			" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
			"  Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR\r\n",
		expPack: "v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
			"\th=from:to:subject:date;\r\n" +
			"\tbh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
			"\tb=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR;\r\n" +
			"\tt=1117574938; x=1577811600; c=simple;\r\n" +
			"\tz=From:foo@eng.example.net|\r\n" +
			"\t To:joe@example.com|\r\n" +
			"\t Subject:demo=20run|\r\n" +
			"\t Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
			"\ti=@eng.example.net; q=dns/txt;\r\n",
		expValidateErr: "dkim: signature is expired at '2005-06-05 21:28:58 +0000 UTC'",
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
		expSimple: "v=1; a=rsa-sha256; s=brisbane; d=example.com;\r\n" +
			"\tc=simple/simple; q=dns/txt; i=joe@football.example.com;\r\n" +
			"\th=Received : From : To : Subject : Date : Message-ID;\r\n" +
			"\tbh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;\r\n" +
			"\tb=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB\r\n" +
			"\t4nujc7YopdG5dWLSdNg6xNAZpOPr+kHxt1IrE+NahM6L/LbvaHut\r\n" +
			"\tKVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrIx0orEtZV\r\n" +
			"\t4bmp/YzhwvcubU4=;\r\n",
		expPack: "v=1; a=rsa-sha256; d=example.com; s=brisbane;\r\n" +
			"\th=received:from:to:subject:date:message-id;\r\n" +
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
		if sig == nil {
			continue
		}

		test.Assert(t, "Signature.Pack", c.expPack, string(sig.Pack()), true)
		test.Assert(t, "Signature.Simple", c.expSimple, string(sig.Simple()), true)

		err = sig.Validate()
		if err != nil {
			test.Assert(t, "Validate error", c.expValidateErr, err.Error(), true)
		}
	}
}

func TestSignatureValidate(t *testing.T) {
	signAlg := SignAlgRS256
	canonSimple := CanonSimple

	cases := []struct {
		desc   string
		sig    *Signature
		exp    string
		expErr string
	}{{
		desc:   "With invalid version",
		sig:    &Signature{},
		expErr: "dkim: invalid version: ''",
	}, {
		desc:   "With empty signature algorithm",
		sig:    &Signature{Version: []byte("1")},
		expErr: errEmptySignAlg.Error(),
	}, {
		desc: "With empty SDID",
		sig: &Signature{
			Version: []byte("1"),
			Alg:     &signAlg,
		},
		expErr: errEmptySDID.Error(),
	}, {
		desc: "With empty selector",
		sig: &Signature{
			Version: []byte("1"),
			Alg:     &signAlg,
			SDID:    []byte("test.com"),
		},
		expErr: errEmptySelector.Error(),
	}, {
		desc: "With empty header",
		sig: &Signature{
			Version:  []byte("1"),
			Alg:      &signAlg,
			SDID:     []byte("test.com"),
			Selector: []byte("mail"),
		},
		expErr: errEmptyHeader.Error(),
	}, {
		desc: "With no From field on header",
		sig: &Signature{
			Version:  []byte("1"),
			Alg:      &signAlg,
			SDID:     []byte("test.com"),
			Selector: []byte("mail"),
			Headers: [][]byte{
				[]byte("to"),
			},
		},
		expErr: errFromHeader.Error(),
	}, {
		desc: "With empty body hash",
		sig: &Signature{
			Version:  []byte("1"),
			Alg:      &signAlg,
			SDID:     []byte("test.com"),
			Selector: []byte("mail"),
			Headers: [][]byte{
				[]byte("from"),
			},
		},
		expErr: errEmptyBodyHash.Error(),
	}, {
		desc: "With empty signature",
		sig: &Signature{
			Version:  []byte("1"),
			Alg:      &signAlg,
			SDID:     []byte("test.com"),
			Selector: []byte("mail"),
			Headers: [][]byte{
				[]byte("from"),
			},
			BodyHash: []byte("bh"),
		},
		expErr: errEmptySignature.Error(),
	}, {
		desc: "With signature create time > expiration time",
		sig: &Signature{
			Version:  []byte("1"),
			Alg:      &signAlg,
			SDID:     []byte("test.com"),
			Selector: []byte("mail"),
			Headers: [][]byte{
				[]byte("from"),
			},
			BodyHash:  []byte("bh"),
			Value:     []byte("b"),
			CreatedAt: 100,
			ExpiredAt: 99,
		},
		expErr: errCreatedTime.Error(),
	}, {
		desc: "With complete signature",
		sig: &Signature{
			Version:  []byte("1"),
			Alg:      &signAlg,
			SDID:     []byte("test.com"),
			Selector: []byte("mail"),
			Headers: [][]byte{
				[]byte("from"),
			},
			BodyHash:    []byte("bh"),
			Value:       []byte("b"),
			CreatedAt:   1000,
			ExpiredAt:   1577811600,
			CanonHeader: &canonSimple,
			PresentHeaders: [][]byte{
				[]byte("Sender:me@domain.com"),
			},
			AUID: []byte("my@test.com"),
		},
		exp: "v=1; a=rsa-sha256; d=test.com; s=mail;\r\n\t" +
			"h=from;\r\n\tbh=bh;\r\n\tb=b;\r\n\t" +
			"t=1000; x=1577811600; c=simple;\r\n\t" +
			"z=Sender:me@domain.com;\r\n\t" +
			"i=my@test.com; \r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := c.sig.Validate()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		got := c.sig.Simple()

		test.Assert(t, "Signature", c.exp, string(got), true)
	}
}

func TestSignatureSet(t *testing.T) {
	cases := []struct {
		desc      string
		t         *tag
		expSimple string
		expErr    string
	}{{
		desc: "With empty input",
		expSimple: "v=; a=rsa-sha256; d=; s=;\r\n\t" +
			"h=;\r\n\tbh=;\r\n\tb=;\r\n\t\r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig := &Signature{}
		err := sig.set(c.t)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		got := sig.Simple()

		test.Assert(t, "Signature", c.expSimple, string(got), true)
	}
}
