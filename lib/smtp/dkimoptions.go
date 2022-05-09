// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"crypto/rsa"

	"github.com/shuLhan/share/lib/email/dkim"
)

// DKIMOptions contains the DKIM signature fields and private key to sign the
// incoming message.
type DKIMOptions struct {
	Signature  *dkim.Signature
	PrivateKey *rsa.PrivateKey
}
