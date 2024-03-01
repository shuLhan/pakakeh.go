// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"crypto/rsa"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestSignatureParse(t *testing.T) {
	cases := []struct {
		desc           string
		in             string
		expErr         string
		expRelaxed     string
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
		expRelaxed:     "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; \r\n",
		expSimple:      "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=/;\r\n",
		expValidateErr: "dkim: invalid version: '2'",
	}, {
		desc:       "With invalid query type",
		in:         "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=s/;\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; \r\n",
		expSimple:  "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=s/;\r\n",
	}, {
		desc:       "With invalid query option",
		in:         "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/relax;\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; \r\n",
		expSimple:  "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/relax;\r\n",
	}, {
		desc:       "Without query type",
		in:         "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns;\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/txt\r\n",
		expSimple:  "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns;\r\n",
	}, {
		desc:       "With empty query type",
		in:         "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/;\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/txt\r\n",
		expSimple:  "v=1; a=rsa-sha256; d=x; s=s; h=from; bh=bh; b=h; q=dns/;\r\n",
	}, {
		desc:           "With unknown tag",
		in:             "v=1;\r\n j=unknown;\r\n a=rsa-sha256; l=512\r\n",
		expRelaxed:     "v=1; a=rsa-sha256; d=; s=; h=; bh=; b=; l=512; \r\n",
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
		expRelaxed: "v=1; a=rsa-sha256; d=example.net; s=brisbane;" +
			" h=from:to:subject:date;" +
			" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;" +
			" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR;" +
			" t=1117574938; x=1577811600; c=simple;" +
			" z=From:foo@eng.example.net|" +
			" To:joe@example.com|" +
			" Subject:demo=20run|" +
			" Date:July=205,=202005=203:44:08=20PM=20-0700;" +
			" i=@eng.example.net; q=dns/txt\r\n",
		expErr:         "dkim: signature is expired at '2019-12-31 17:00:00 +0000 UTC'",
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
		expRelaxed: "v=1; a=rsa-sha256; d=example.com; s=brisbane;" +
			" h=received:from:to:subject:date:message-id;" +
			" bh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;" +
			" b=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB" +
			"4nujc7YopdG5dWLSdNg6xNAZpOPr+kHxt1IrE+NahM6L/LbvaHut" +
			"KVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrIx0orEtZV" +
			"4bmp/YzhwvcubU4=;" +
			" c=simple/simple;" +
			" i=joe@football.example.com; q=dns/txt\r\n",
	}, {
		desc:       `With expired-at more than 12 digits`,
		in:         "v=1; a=rsa-sha256; d=example.com; s=mykey; h=from; bh=bh; b=b; t=1117574938; x=9223372036854775808\r\n",
		expRelaxed: "v=1; a=rsa-sha256; d=example.com; s=mykey; h=from; bh=bh; b=b; t=1117574938; x=9223372036854775807; \r\n",
		expSimple:  "v=1; a=rsa-sha256; d=example.com; s=mykey; h=from; bh=bh; b=b; t=1117574938; x=9223372036854775808\r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig, err := Parse([]byte(c.in))
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
		if sig == nil {
			continue
		}

		test.Assert(t, "Signature.Relaxed", c.expRelaxed, string(sig.Relaxed()))
		test.Assert(t, "Signature.Simple", c.expSimple, string(sig.Simple()))

		err = sig.Validate()
		if err != nil {
			test.Assert(t, "Validate error", c.expValidateErr, err.Error())
		}
	}
}

func TestNewSignature(t *testing.T) {
	cases := []struct {
		sdid     string
		selector string
		exp      string
	}{{
		sdid:     "",
		selector: "",
		exp:      "v=1; a=rsa-sha256; d=; s=; h=; bh=; b=; c=relaxed/relaxed; \r\n",
	}, {
		sdid:     "d",
		selector: "s",
		exp:      "v=1; a=rsa-sha256; d=d; s=s; h=; bh=; b=; c=relaxed/relaxed; \r\n",
	}}

	for _, c := range cases {
		t.Log(c.sdid + " " + c.selector)

		got := NewSignature([]byte(c.sdid), []byte(c.selector))

		test.Assert(t, "Signature", c.exp, string(got.Relaxed()))
	}
}

func TestSignatureHash(t *testing.T) {
	sig := &Signature{}

	cases := []struct {
		desc string
		in   string
		exp  string
		alg  SignAlg
	}{{
		desc: "With empty input, sha1",
		alg:  SignAlgRS1,
		exp:  "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
	}, {
		desc: "With CRLF, sha1",
		alg:  SignAlgRS1,
		in:   "\r\n",
		exp:  "uoq1oCgLlTqpdDX/iUbLy7J1Wic=",
	}, {
		desc: "With empty input, sha256",
		alg:  SignAlgRS256,
		exp:  "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
	}, {
		desc: "With CRLF, sha256",
		alg:  SignAlgRS256,
		in:   "\r\n",
		exp:  "frcCV1k9oG9oKj3dpUqdJg1PxRT2RSN/XKdLCPjaYaY=",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig.Alg = &c.alg

		_, got64 := sig.Hash([]byte(c.in))

		test.Assert(t, "Hash", c.exp, string(got64))
	}
}

func TestSignatureSign(t *testing.T) {
	if privateKey == nil {
		initKeys(t)
	}

	sig := &Signature{}

	cases := []struct {
		pk      *rsa.PrivateKey
		desc    string
		exp     string
		expErr  string
		input   []byte
		hashAlg SignAlg
		signAlg SignAlg
	}{{
		desc:   "With empty private key",
		expErr: "email/dkim: empty private key for signing",
	}, {
		desc:    "With failed signing",
		pk:      privateKey,
		hashAlg: SignAlgRS1,
		signAlg: SignAlgRS256,
		expErr:  "email/dkim: failed to sign message: crypto/rsa: input must be hashed message",
	}, {
		desc:    "With empty hash, sha1",
		pk:      privateKey,
		hashAlg: SignAlgRS1,
		signAlg: SignAlgRS1,
		exp:     "jPS+9gKtcUOlO5ITvvu0vRW9WnroERAqgGQuzsN7yrfdK5qTcT1UFtF1Mdkz3vu0YVDItX5UrgCVEM8sqwy2g7CfWIbNTKgxfkGZCnPUiVmbJcM6TvMoGqmbPNMDN8VGibfp6uVEm5bxPSFBZvrc5OX2fqvUt0NfQYHiUluYdY8=",
	}, {
		desc:    "With CRLF, sha1",
		pk:      privateKey,
		hashAlg: SignAlgRS1,
		signAlg: SignAlgRS1,
		input:   []byte("\r\n"),
		exp:     "5XAKcEb55b7wZrOzrKNjykgRXdloRIdtkuXIQ36Ux9/G2QHJ+kHT7uGY5IjhzSvmYHwKN+tPy1iSZl35XuPWlcupyU+1h+5rwrrvF6JGz1HBVEP23oCzM82opIJF80Dde9YDHH1/Q62WVcM3zAgp9MoCJeO7JToHKKTU/OKXmvo=",
	}, {
		desc:    "With empty hash, sha256",
		pk:      privateKey,
		hashAlg: SignAlgRS256,
		signAlg: SignAlgRS256,
		exp:     "X5VCoHS9jdAuk/fujKm9SNJQPQOHnngURIrTnTDGoxpFzwuejsVYs+QqNi+gch4st5rpDklulCxrIT6grT+XEo6nYnUUs/i5cOAnsWHw+1jg15GSk37eHKZDYW9qHgxpZqiPQjorDfCvLsKWpb5DKzelkkB6lerS3amBuv8gkkw=",
	}, {
		desc:    "With CRLF, sha256",
		pk:      privateKey,
		hashAlg: SignAlgRS256,
		signAlg: SignAlgRS256,
		input:   []byte("\r\n"),
		exp:     "LwMB2vdR7qDcHr8VS758WUtTECOrwAIlS9eRZUoEGP1SCpl5RzVJ1mmMD7bq72djTQ3loMA4JyBJSm/PUahkECuWuyCu+LkkX4QcoosrWJj01cNA9SG3VuDBoDELbf7rR9Z9h7ObWCTzodrWsCeg0tRpI1Z2AM9mRJWSVBEjdKI=",
	}, {
		desc:    "With text, sha256",
		pk:      privateKey,
		hashAlg: SignAlgRS256,
		signAlg: SignAlgRS256,
		input:   []byte("text"),
		exp:     "wJOt/N8VsUR4dczLN/8MqxMBXgDyl0lS8AC7sJYSukbrqO2hhIeNcBHccx2sWo/CGVPWton7DNzQfPv56y0kXjlrDOzZCSU3sqEb81S7n4BYkLuBnOoWsQKZrUr/PnuUGS48/Jz/c+X99y4iFx0myOI0iHCGK47uaQE/XNwUBXs=",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig.Alg = &c.hashAlg

		hashed, _ := sig.Hash(c.input)

		sig.Alg = &c.signAlg

		err := sig.Sign(c.pk, hashed)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Signature", c.exp, string(sig.Value))
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
		exp: "v=1; a=rsa-sha256; d=test.com; s=mail;\r\n " +
			"h=from;\r\n bh=bh;\r\n b=b;\r\n " +
			"t=1000; x=1577811600; c=simple;\r\n " +
			"z=Sender:me@domain.com;\r\n " +
			"i=my@test.com; \r\n",
		expErr: "dkim: signature is expired at '2019-12-31 17:00:00 +0000 UTC'",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := c.sig.Validate()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		got := c.sig.Simple()

		test.Assert(t, "Signature", c.exp, string(got))
	}
}

func TestSignatureVerify(t *testing.T) {
	if publicKey == nil {
		initKeys(t)
	}

	sig := &Signature{}

	cases := []struct {
		key      *Key
		desc     string
		sigValue string
		input    string
		expErr   string
		sigAlg   SignAlg
	}{{
		desc:   "With empty key",
		expErr: "email/dkim: key record is empty",
	}, {
		desc:   "With empty public key",
		key:    &Key{},
		expErr: "email/dkim: public key is empty",
	}, {
		desc:     "With invalid base64",
		sigAlg:   SignAlgRS1,
		sigValue: "invalidBASE64(:)",
		key: &Key{
			RSA: publicKey,
		},
		expErr: "email/dkim: failed to decode signature: illegal base64 data at input byte 13",
	}, {
		desc:   "With empty signature",
		sigAlg: SignAlgRS256,
		key: &Key{
			RSA: publicKey,
		},
		input:  "",
		expErr: "email/dkim: verification failed: crypto/rsa: verification error",
	}, {
		desc:     "With CRLF, sha1",
		sigAlg:   SignAlgRS1,
		sigValue: "5XAKcEb55b7wZrOzrKNjykgRXdloRIdtkuXIQ36Ux9/G2QHJ+kHT7uGY5IjhzSvmYHwKN+tPy1iSZl35XuPWlcupyU+1h+5rwrrvF6JGz1HBVEP23oCzM82opIJF80Dde9YDHH1/Q62WVcM3zAgp9MoCJeO7JToHKKTU/OKXmvo=",
		key: &Key{
			RSA: publicKey,
		},
		input: "\r\n",
	}, {
		desc:     "With CRLF, sha256",
		sigAlg:   SignAlgRS256,
		sigValue: "LwMB2vdR7qDcHr8VS758WUtTECOrwAIlS9eRZUoEGP1SCpl5RzVJ1mmMD7bq72djTQ3loMA4JyBJSm/PUahkECuWuyCu+LkkX4QcoosrWJj01cNA9SG3VuDBoDELbf7rR9Z9h7ObWCTzodrWsCeg0tRpI1Z2AM9mRJWSVBEjdKI=",
		key: &Key{
			RSA: publicKey,
		},
		input: "\r\n",
	}, {
		desc:     "With text, sha256",
		sigAlg:   SignAlgRS256,
		sigValue: "wJOt/N8VsUR4dczLN/8MqxMBXgDyl0lS8AC7sJYSukbrqO2hhIeNcBHccx2sWo/CGVPWton7DNzQfPv56y0kXjlrDOzZCSU3sqEb81S7n4BYkLuBnOoWsQKZrUr/PnuUGS48/Jz/c+X99y4iFx0myOI0iHCGK47uaQE/XNwUBXs=",
		key: &Key{
			RSA: publicKey,
		},
		input: "text",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig.Alg = &c.sigAlg
		sig.Value = []byte(c.sigValue)

		bhash, _ := sig.Hash([]byte(c.input))

		err := sig.Verify(c.key, bhash)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
		}
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
		expSimple: "v=; a=rsa-sha256; d=; s=;\r\n " +
			"h=;\r\n bh=;\r\n b=;\r\n \r\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sig := &Signature{}
		err := sig.set(c.t)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		got := sig.Simple()

		test.Assert(t, "Signature", c.expSimple, string(got))
	}
}
