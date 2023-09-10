// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import "crypto/ed25519"

type Key struct {
	// AllowedSubjects contains list of subject that are allowed in the
	// token's claim "sub" to be signed by this public key.
	// This field is used by receiver to check the claim "sub" and compare
	// it with this list.
	// Empty list means allowing all subjects.
	AllowedSubjects map[string]struct{}

	// ID is a unique key ID.
	ID string

	// PrivateKey for signing public token.
	Private ed25519.PrivateKey

	// PublicKey for verifying public token.
	Public ed25519.PublicKey
}
