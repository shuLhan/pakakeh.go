// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/email/dkim"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func TestMain(m *testing.M) {
	Epoch = func() int64 {
		return 1645811431
	}

	os.Exit(m.Run())
}

func initKeys(t *testing.T) {
	rsaPrivateRaw, err := os.ReadFile("dkim/testdata/rsa.private.pem")
	if err != nil {
		t.Fatal(err)
	}
	rsaPublicRaw, err := os.ReadFile("dkim/testdata/rsa.public.pem")
	if err != nil {
		t.Fatal(err)
	}

	rsaPrivatePEM, _ := pem.Decode(rsaPrivateRaw)
	rsaPublicPEM, _ := pem.Decode(rsaPublicRaw)

	privateKey, err = x509.ParsePKCS1PrivateKey(rsaPrivatePEM.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	ipublicKey, err := x509.ParsePKIXPublicKey(rsaPublicPEM.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	publicKey = ipublicKey.(*rsa.PublicKey)

	key := &dkim.Key{
		RSA:       publicKey,
		ExpiredAt: time.Now().Add(time.Hour).Unix(),
	}

	dname := "brisbane._domainkey.example.com"
	dkim.DefaultKeyPool.Put(dname, key)
}

func TestSanitize(t *testing.T) {
	type testCase struct {
		in  []byte
		exp []byte
	}

	var cases = []testCase{{
		in:  []byte("not\n a\t comment"),
		exp: []byte("not a comment"),
	}, {
		in:  []byte("A B \n (comment \t) C \r\n ( \tcomment )\r\n\tD\r\n "),
		exp: []byte(`A B C D`),
	}, {
		in:  []byte("A B \r\n ( C (D\r\n (E)) \t) F\r\n "),
		exp: []byte(`A B F`),
	}, {
		in:  []byte("Fri, 21 Nov 1997 09(comment): 55 : 06 -0600\r\n"),
		exp: []byte("Fri, 21 Nov 1997 09: 55 : 06 -0600"),
	}}

	var (
		c   testCase
		got []byte
	)
	for _, c = range cases {
		got = sanitize(c.in)
		test.Assert(t, `sanitize`, string(c.exp), string(got))
	}
}
