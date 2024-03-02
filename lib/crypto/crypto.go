// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crypto provide a wrapper for standard crypto package and
// golang.org/x/crypto.
package crypto

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/os/exec"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// ErrEmptyPassphrase returned when private key is encrypted and loaded
// interactively, using [LoadPrivateKeyInteractive], but the readed
// passphrase is empty.
//
// This is to catch error "bcrypt_pbkdf: empty password" earlier that cannot
// be catched using [errors.Is] after
// [ssh.ParseRawPrivateKeyWithPassphrase].
var ErrEmptyPassphrase = errors.New(`empty passphrase`)

// ErrStdinPassphrase error when program cannot changes [os.Stdin] for
// reading passphrase in terminal.
// The original error message is "inappropriate ioctl for device".
var ErrStdinPassphrase = errors.New(`cannot read passhprase from stdin`)

// List of environment variables reads when reading passphrase
// interactively.
const (
	envKeySSHAskpassRequire = `SSH_ASKPASS_REQUIRE` //nolint: gosec
	envKeySSHAskpass        = `SSH_ASKPASS`
	envKeyDisplay           = `DISPLAY`
)

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
// from terminal or from program defined in SSH_ASKPASS environment
// variable.
//
// The termrw parameter is optional, default to os.Stdin if its nil.
// Its provide as reader-and-writer to prompt and read password from
// terminal (or for testing).
//
// The SSH_ASKPASS is controlled by environment SSH_ASKPASS_REQUIRE.
//
//   - If SSH_ASKPASS_REQUIRE is empty the passphrase will read from
//     terminal first, if not possible then using SSH_ASKPASS program.
//
//   - If SSH_ASKPASS_REQUIRE is set to "never", the passphrase will read
//     from terminal only.
//
//   - If SSH_ASKPASS_REQUIRE is set to "prefer", the passphrase will read
//     using SSH_ASKPASS program not from terminal, but require
//     DISPLAY environment to be set.
//
//   - If SSH_ASKPASS_REQUIRE is set to "force", the passphrase will read
//     using SSH_ASKPASS program not from terminal without checking DISPLAY
//     environment.
//
// See ssh(1) manual page for more information.
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
		askpassRequire = os.Getenv(envKeySSHAskpassRequire)

		pass string
	)

	switch askpassRequire {
	default:
		// Accept empty SSH_ASKPASS_REQUIRE, "never", or other
		// unknown string.
		// Try to read passphrase from terminal first.
		pass, err = readPassTerm(termrw, file)
		if err != nil {
			if !strings.Contains(err.Error(), `inappropriate ioctl`) {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}
			if askpassRequire == `` {
				// We cannot changes the os.Stdin to raw
				// terminal, try using SSH_ASKPASS program
				// instead.
				pass, err = sshAskpass(askpassRequire)
				if err != nil {
					return nil, fmt.Errorf(`%s: %w`, logp, err)
				}
			}
		}

	case `prefer`, `force`:
		// Use SSH_ASKPASS program instead of reading from terminal.
		pass, err = sshAskpass(askpassRequire)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	passphrase = []byte(pass)

	pkey, err = ssh.ParseRawPrivateKeyWithPassphrase(rawpem, passphrase)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return pkey, nil
}

// readPassTerm read the passphrase from terminal.
func readPassTerm(termrw io.ReadWriter, file string) (pass string, err error) {
	var (
		prompt = fmt.Sprintf(`Enter passphrase for %q:`, file)

		xterm *term.Terminal
	)

	if termrw == nil {
		var (
			stdin = int(os.Stdin.Fd())

			oldState *term.State
		)

		oldState, err = term.MakeRaw(stdin)
		if err != nil {
			return ``, fmt.Errorf(`MakeRaw: %w`, err)
		}
		defer func() {
			_ = term.Restore(stdin, oldState)
		}()

		termrw = os.Stdin
	}

	xterm = term.NewTerminal(termrw, ``)

	pass, err = xterm.ReadPassword(prompt)
	if err != nil && !errors.Is(err, io.EOF) {
		return ``, fmt.Errorf(`ReadPassword: %w`, err)
	}
	if len(pass) == 0 {
		return ``, ErrEmptyPassphrase
	}
	return pass, nil
}

// sshAskpass get passphrase from the program defined in environment
// SSH_ASKPASS.
// This require the DISPLAY environment also be set only if
// SSH_ASKPASS_REQUIRE is set to "prefer".
func sshAskpass(askpassRequire string) (pass string, err error) {
	var val string

	if askpassRequire == `prefer` {
		val = os.Getenv(envKeyDisplay)
		if len(val) == 0 {
			return ``, ErrStdinPassphrase
		}
	}

	val = os.Getenv(envKeySSHAskpass)
	if len(val) == 0 {
		return ``, ErrStdinPassphrase
	}

	var stdout bytes.Buffer

	err = exec.Run(val, &stdout, os.Stderr)
	if err != nil {
		return ``, err
	}

	pass = strings.TrimSpace(stdout.String())

	return pass, nil
}
