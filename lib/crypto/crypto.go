// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crypto provide a wrapper for standard crypto package and
// golang.org/x/crypto.
package crypto

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// ErrEmptyPassphrase returned when private key is encrypted and loaded
// interactively, using [LoadPrivateKeyInteractive], but the readed
// passphrase is empty from terminal.
//
// This is to catch error "bcrypt_pbkdf: empty password" earlier that cannot
// be catched using errors.Is after [ssh.ParseRawPrivateKeyWithPassphrase].
var ErrEmptyPassphrase = errors.New(`empty passphrase`)

// DecryptOaep extend the [rsa.DecryptOAEP] to make it able to decrypt a
// message larger than its public modulus size.
func DecryptOaep(hash hash.Hash, random io.Reader, pkey *rsa.PrivateKey, cipher, label []byte) (plain []byte, err error) {
	var (
		logp   = `DecryptOaep`
		msglen = len(cipher)
		limit  = pkey.PublicKey.Size()

		chunkPlain  []byte
		chunkCipher []byte
		x           int
	)
	for x < msglen {
		if x+limit > msglen {
			chunkCipher = cipher[x:]
		} else {
			chunkCipher = cipher[x : x+limit]
		}
		chunkPlain, err = rsa.DecryptOAEP(hash, random, pkey, chunkCipher, label)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		plain = append(plain, chunkPlain...)
		x += limit
	}
	return plain, nil
}

// EncryptOaep extend the [rsa.EncryptOAEP] to make it able to encrypt a
// message larger than its than (public modulus size - 2*hash.Size - 2).
//
// The function signature is the same with [rsa.EncryptOAEP] except the
// name, to make it distinguishable.
func EncryptOaep(hash hash.Hash, random io.Reader, pub *rsa.PublicKey, msg, label []byte) (cipher []byte, err error) {
	var (
		logp   = `EncryptOaep`
		msglen = len(msg)
		limit  = pub.Size() - 2*hash.Size() - 2

		chunkPlain  []byte
		chunkCipher []byte
		x           int
	)
	for x < msglen {
		if x+limit > msglen {
			chunkPlain = msg[x:]
		} else {
			chunkPlain = msg[x : x+limit]
		}
		chunkCipher, err = rsa.EncryptOAEP(hash, random, pub, chunkPlain, label)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		cipher = append(cipher, chunkCipher...)
		x += limit
	}
	return cipher, nil
}

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

// LoadPrivateKeyInteractive load the private key from file.
// If the private key file is encrypted, it will prompt for the passphrase
// from terminal.
//
// The termrw parameter is optional, default to os.Stdin if its nil.
// Its provide as reader-and-writer to prompt and read password from
// terminal; or for testing.
func LoadPrivateKeyInteractive(termrw io.ReadWriter, file string) (pkey crypto.PrivateKey, err error) {
	var (
		logp = `LoadPrivateKeyInteractive`

		passphrase []byte
		rawpem     []byte
	)

	rawpem, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	pkey, err = ssh.ParseRawPrivateKey(rawpem)
	if err == nil {
		return pkey, nil
	}

	// We can use "err.(*ssh.PassphraseMissingError)", but I don't trust
	// the golang.org/x/crypto to return error as is yet.
	if !strings.Contains(err.Error(), `passphrase protected`) {
		return nil, err
	}

	var (
		prompt = fmt.Sprintf(`Enter passphrase for %q:`, file)

		xterm *term.Terminal
		pass  string
	)

	if termrw == nil {
		var (
			stdin = int(os.Stdin.Fd())

			oldState *term.State
		)

		oldState, err = term.MakeRaw(stdin)
		if err != nil {
			return nil, fmt.Errorf(`%s: MakeRaw: %w`, logp, err)
		}
		defer term.Restore(stdin, oldState)

		termrw = os.Stdin
	}

	xterm = term.NewTerminal(termrw, ``)

	pass, err = xterm.ReadPassword(prompt)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf(`%s: ReadPassword: %w`, logp, err)
	}
	if len(pass) == 0 {
		return nil, fmt.Errorf(`%s: %w`, logp, ErrEmptyPassphrase)
	}

	passphrase = []byte(pass)

	pkey, err = ssh.ParseRawPrivateKeyWithPassphrase(rawpem, passphrase)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return pkey, nil
}
