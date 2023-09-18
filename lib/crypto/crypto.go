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
//
// The passphrase is optional and will only be used if the private key is
// encrypted.
// If its set it will use [ssh.ParseRawPrivateKeyWithPassphrase].
func LoadPrivateKey(file string, passphrase []byte) (pkey crypto.PrivateKey, err error) {
	var (
		logp = `LoadPrivateKey`

		rawpem []byte
	)

	rawpem, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	if len(passphrase) != 0 {
		pkey, err = ssh.ParseRawPrivateKeyWithPassphrase(rawpem, passphrase)
	} else {
		pkey, err = ssh.ParseRawPrivateKey(rawpem)
	}
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return pkey, nil
}
