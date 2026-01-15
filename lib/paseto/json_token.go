// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"fmt"
	"time"
)

const (
	_validateTimeDrift = 5 * time.Second
)

// JSONToken define the metadata and data inside the token that are included
// to generate the signature.
type JSONToken struct {
	Issuer    string     `json:"iss,omitempty"`
	Subject   string     `json:"sub,omitempty"`
	Audience  string     `json:"aud,omitempty"`
	ExpiredAt *time.Time `json:"exp,omitempty"`
	NotBefore *time.Time `json:"nbf,omitempty"`
	IssuedAt  *time.Time `json:"iat,omitempty"`
	TokenID   string     `json:"jti,omitempty"`
	Data      string     `json:"data"`
}

// Validate the JSON token fields,
//
//   - The Issuer must equal to peer.ID
//   - The Audience must equal to received ID,
//   - If peer.AllowedSubjects is not empty, the Subject value must be in
//     one of them,
//   - The current time must be after IssuedAt field,
//   - The current time must after NotBefore "nbf" field,
//   - The current time must before ExpiredAt field.
//
// If one of the above condition is not passed, it will return an error.
func (jtoken *JSONToken) Validate(audience string, peer Key) (err error) {
	now := time.Now().Round(time.Second)
	if jtoken.Issuer != peer.ID {
		return fmt.Errorf("expecting issuer %q, got %q", peer.ID,
			jtoken.Issuer)
	}
	if len(peer.AllowedSubjects) != 0 {
		_, isAllowed := peer.AllowedSubjects[jtoken.Subject]
		if !isAllowed {
			return fmt.Errorf("token subject %q is not allowed for key %q",
				jtoken.Subject, peer.ID)
		}
	}
	if len(audience) != 0 {
		if jtoken.Audience != audience {
			return fmt.Errorf("expecting audience %q, got %q",
				audience, jtoken.Audience)
		}
	}
	if jtoken.IssuedAt != nil {
		diff := now.Sub(*jtoken.IssuedAt)
		if diff < 0 {
			diff *= -1
		}
		if diff > _validateTimeDrift {
			return fmt.Errorf("token issued at %s before current time %s",
				jtoken.IssuedAt, now)
		}
	}
	if jtoken.NotBefore != nil {
		diff := now.Sub(*jtoken.NotBefore)
		if diff < 0 {
			diff *= -1
		}
		if diff > _validateTimeDrift {
			return fmt.Errorf("token must not used before %s", jtoken.NotBefore)
		}
	}
	if jtoken.ExpiredAt != nil {
		if now.After(*jtoken.ExpiredAt) {
			return fmt.Errorf("token is expired %s", jtoken.ExpiredAt)
		}
	}
	return nil
}
