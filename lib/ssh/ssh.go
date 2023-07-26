// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ssh provide a wrapper for golang.org/x/crypto/ssh and a parser for
// SSH client configuration specification ssh_config(5).
package ssh

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// LoadPrivateKeyInteractive load private key from file.
// If key is encrypted, it will prompt the passphrase in terminal with
// maximum maxAttempt times.
// If the passphrase still invalid after maxAttempt it will return an error.
func LoadPrivateKeyInteractive(file string, maxAttempt int) (pkey any, err error) {
	var (
		logp = `LoadPrivateKeyInteractive`

		pkeyPem       []byte
		pass          []byte
		isMissingPass bool
	)

	pkeyPem, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	pkey, err = ssh.ParseRawPrivateKey(pkeyPem)
	if err == nil {
		return pkey, nil
	}

	_, isMissingPass = err.(*ssh.PassphraseMissingError)
	if !isMissingPass {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	for x := 0; x < maxAttempt; x++ {
		fmt.Printf(`Enter passphrase for %q:`, file)
		pass, err = term.ReadPassword(0)
		fmt.Println(``)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		pkey, err = ssh.ParseRawPrivateKeyWithPassphrase(pkeyPem, pass)
		if err != nil {
			log.Printf(`%s: %s`, logp, err)
			continue
		}
		break
	}
	if pkey == nil {
		// Invalid passphrase after three attempts.
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return pkey, nil
}
