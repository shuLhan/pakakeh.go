// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"
)

// LocalMode implement the PASETO encrypt and decrypt using shared key.
type LocalMode struct {
	aead cipher.AEAD
}

// NewLocalMode create and initialize new LocalMode using shared key.
func NewLocalMode(key []byte) (local *LocalMode, err error) {
	local = &LocalMode{}

	local.aead, err = chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	return local, nil
}

// Pack encrypt the data and generate token with optional footer.
func (l *LocalMode) Pack(data, footer []byte) (token string, err error) {
	return Encrypt(l.aead, data, footer)
}

// Unpack decrypt the token and return the plain data and optional footer.
func (l *LocalMode) Unpack(token string) (data, footer []byte, err error) {
	return Decrypt(l.aead, token)
}
