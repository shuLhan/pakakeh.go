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
const DriftSeconds = 5 * time.Second

// List of errors for [Payload.Validate].
var (
	ErrUnknownIssuer   = errors.New(`unknown issuer`)
	ErrUnknownSubject  = errors.New(`unknown subject`)
	ErrInvalidAudience = errors.New(`invalid audience`)
	ErrExpired         = errors.New(`expired`)
	ErrNotBeforeAfter  = errors.New(`not-before is after current time`)
)

// Payload represents the data and claims.
//
// The claims follow RFC 7519 that includes issuer, subject, audience,
// expiration time, not-before time, issued-at, and ID.
type Payload struct {
	// Issuer defines the peer ID that issued the payload.
	Issuer string `json:"iss,omitempty"`

	// Subject defines the scope of payload.
	Subject string `json:"sub,omitempty"`

	// Audience defines the peer ID that receive the payload.
	Audience string `json:"aud,omitempty"`

	// ExpiredAt defines the expiration time when the payload MUST NOT
	// be accepted for processing.
	ExpiredAt *time.Time `json:"exp,omitempty"`

	// NotBefore defines the time when the payload MUST NOT be accepted
	// for processing.
	NotBefore *time.Time `json:"nbf,omitempty"`

	// IssuedAt defines the time at which the payload is issued.
	IssuedAt *time.Time `json:"iat,omitempty"`

	// TokenID defines the unique identifier for the payload.
	TokenID string `json:"jti,omitempty"`

	// Data defines actual information to be send in message.
	// Data must be JSON encodable.
	Data any `json:"data"`
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

	now := time.Now().Round(time.Second)
	if pload.Issuer != sender.ID {
		return fmt.Errorf(`%s: %w`, logp, ErrUnknownIssuer)
	}
	if len(sender.AllowedSubjects) != 0 {
		_, ok := sender.AllowedSubjects[pload.Subject]
		if !ok {
			return fmt.Errorf(`%s: %w`, logp, ErrUnknownSubject)
		}
	}
	if len(recvID) != 0 && pload.Audience != recvID {
		return fmt.Errorf(`%s: %w`, logp, ErrInvalidAudience)
	}
	if pload.ExpiredAt != nil {
		diff := pload.ExpiredAt.Sub(now)
		diff -= DefaultTTL
		if diff < 0 {
			diff *= -1
		}
		if diff > DriftSeconds {
			return fmt.Errorf(`%s: %w`, logp, ErrExpired)
		}
	}
	if pload.NotBefore != nil {
		diff := now.Sub(*pload.NotBefore)
		if diff < 0 {
			diff *= -1
		}
		if diff > DriftSeconds {
			return fmt.Errorf(`%s: %w`, logp, ErrNotBeforeAfter)
		}
	}
	return nil
}
