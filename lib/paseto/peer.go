// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import "crypto/ed25519"

// Peer represents the principal in public mode, the one that sends and
// receives the message.
type Peer struct {
	// AllowedSubjects list of scope that are allowed for this peer.
	// This field is used by receiver to check the claim "sub" and compare
	// it with this list.
	// Empty list means allowing all scopes.
	AllowedSubjects map[string]struct{}

	// ID is a unique identifier of the peer.
	ID string

	// PrivateKey for signing message.
	Private ed25519.PrivateKey

	// PublicKey for verifying message.
	Public ed25519.PublicKey
}
