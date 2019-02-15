// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/email/dkim"
)

var ( // nolint: gochecknoglobals
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func initKeys(t *testing.T) {
	rsaPrivateRaw, err := ioutil.ReadFile("dkim/testdata/rsa.private.pem")
	if err != nil {
		t.Fatal(err)
	}
	rsaPublicRaw, err := ioutil.ReadFile("dkim/testdata/rsa.public.pem")
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
