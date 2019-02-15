// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"
)

var ( // nolint: gochecknoglobals
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func initKeys(t *testing.T) {
	rsaPrivateRaw, err := ioutil.ReadFile("testdata/rsa.private.pem")
	if err != nil {
		t.Fatal(err)
	}
	rsaPublicRaw, err := ioutil.ReadFile("testdata/rsa.public.pem")
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

	key := &Key{
		RSA:       publicKey,
		ExpiredAt: time.Now().Add(time.Hour).Unix(),
	}

	dname := "brisbane._domainkey.example.com"
	DefaultKeyPool.Put(dname, key)
}
