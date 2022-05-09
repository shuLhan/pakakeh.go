// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crypto provide a wrapper for standard crypto package.
package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPrivateKey read and parse PEM formatted private key from file.
func LoadPrivateKey(file string) (pkey *rsa.PrivateKey, err error) {
	rawPEM, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(rawPEM)
	if block == nil {
		return nil, fmt.Errorf("crypto: failed to parse PEM block from %q", file)
	}

	pkey, err = x509.ParsePKCS1PrivateKey(block.Bytes)

	return
}
