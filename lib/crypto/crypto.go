// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crypto provide a wrapper for standard crypto package and
// golang.org/x/crypto.
package crypto

import (
	"crypto"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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

		passphrase    []byte
		rawpem        []byte
		isMissingPass bool
	)

	rawpem, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	pkey, err = ssh.ParseRawPrivateKey(rawpem)
	if err == nil {
		return pkey, nil
	}

	_, isMissingPass = err.(*ssh.PassphraseMissingError)
	if !isMissingPass {
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

	passphrase = []byte(pass)

	pkey, err = ssh.ParseRawPrivateKeyWithPassphrase(rawpem, passphrase)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return pkey, nil
}
