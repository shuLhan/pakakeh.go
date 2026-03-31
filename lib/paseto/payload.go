// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"errors"
	"fmt"
	"time"
)

// DriftSeconds defines the time differences in seconds allowed in
// [Payload.ExpiredAt] and [Payload.NotBefore].
// This value can be changes to increases or decreases the leeway during
// [Payload.Validate].
var DriftSeconds int64 = 5

// List of errors for [Payload.Validate].
var (
	ErrPayloadAudience  = errors.New(`invalid audience`)
	ErrPayloadExpired   = errors.New(`expired`)
	ErrPayloadIssuer    = errors.New(`unknown issuer`)
	ErrPayloadNotBefore = errors.New(`token cannot be used yet`)
	ErrPayloadSubject   = errors.New(`unknown subject`)
)

// Payload represents the data and claims.
//
// The claims follow RFC 7519 that includes issuer, subject, audience,
// expiration time, not-before time, issued-at, and ID.
type Payload struct {
	// Data defines actual information to be send in message.
	// Data must be JSON encodable.
	Data any `json:"data"`

	// Issuer defines the peer ID that issued the payload.
	Issuer string `json:"iss,omitempty"`

	// Subject defines the scope of payload.
	Subject string `json:"sub,omitempty"`

	// Audience defines the peer ID that receive the payload.
	Audience string `json:"aud,omitempty"`

	// TokenID defines the unique identifier for the payload.
	TokenID string `json:"jti,omitempty"`

	// ExpiredAt defines the expiration time when the payload MUST NOT
	// be accepted for processing.
	ExpiredAt int64 `json:"exp,omitempty"`

	// NotBefore defines the time when the payload MUST NOT be accepted
	// for processing.
	NotBefore int64 `json:"nbf,omitempty"`

	// IssuedAt defines the time at which the payload is issued.
	IssuedAt int64 `json:"iat,omitempty"`
}

// Validate validates the the payload and returns nil when all of the
// following conditions satisfied,
//
//   - The Issuer field MUST equal to sender [Peer.ID].
//   - The Subject field MUST be in one of sender [Peer.AllowedSubjects] if
//     its is not empty.
//   - The Audience field MUST equal to recvID if its not empty.
//   - The ExpiredAt field MUST before the current time.
//   - The NotBefore field MUST equal or before the current time.
//
// The relation between current time, ExpiredAt, and NotBefore can be viewed
// this way
//
//	          iat/nbf             exp
//	        -----o-----------------o----- time
//	            / \               / \
//	invalid ---+   +--- valid ---+   +--- invalid
//
// If one of the above condition is not passed, it will return an error.
func (pload *Payload) Validate(recvID string, sender Peer) (err error) {
	logp := `payload`

	now := time.Now().UTC().Unix()
	if pload.Issuer != sender.ID {
		return fmt.Errorf(`%s: %w`, logp, ErrPayloadIssuer)
	}
	if len(sender.AllowedSubjects) != 0 {
		_, ok := sender.AllowedSubjects[pload.Subject]
		if !ok {
			return fmt.Errorf(`%s: %w: %s`, logp,
				ErrPayloadSubject, pload.Subject)
		}
	}
	if len(recvID) != 0 && pload.Audience != recvID {
		return fmt.Errorf(`%s: %w: %s`, logp, ErrPayloadAudience,
			pload.Audience)
	}
	if pload.ExpiredAt != 0 {
		diff := pload.ExpiredAt - now
		diff -= DefaultTTL
		if diff < 0 {
			diff *= -1
		}
		if diff > DriftSeconds {
			return fmt.Errorf(`%s: %w: exp is %s`, logp,
				ErrPayloadExpired,
				time.Unix(pload.ExpiredAt, 0).UTC())
		}
	}
	if pload.NotBefore != 0 {
		diff := now - pload.NotBefore
		if diff < 0 {
			diff *= -1
		}
		if diff > DriftSeconds {
			return fmt.Errorf(`%s: %w: nbf is %s`, logp,
				ErrPayloadNotBefore,
				time.Unix(pload.NotBefore, 0).UTC())
		}
	}
	return nil
}
