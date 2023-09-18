// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crypto provide a wrapper for standard crypto package and
// golang.org/x/crypto.
package crypto

import (
	"crypto"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// LoadPrivateKey read and parse PEM formatted private key from file.
// This is a wrapper for [ssh.ParseRawPrivate] that can return either
// *dsa.PrivateKey, ecdsa.PrivateKey, *ed25519.PrivateKey, or
// *rsa.PrivateKey.
func LoadPrivateKey(file string) (pkey crypto.PrivateKey, err error) {
	var (
		logp = `LoadPrivateKey`

		rawpem []byte
	)

	rawpem, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	pkey, err = ssh.ParseRawPrivateKey(rawpem)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return
}
