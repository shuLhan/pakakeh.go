// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package smtp

import (
	"crypto/rsa"

	"git.sr.ht/~shulhan/pakakeh.go/lib/email/dkim"
)

// DKIMOptions contains the DKIM signature fields and private key to sign the
// incoming message.
type DKIMOptions struct {
	Signature  *dkim.Signature
	PrivateKey *rsa.PrivateKey
}
