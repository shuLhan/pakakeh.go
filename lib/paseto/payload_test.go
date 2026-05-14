// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"testing"
	"testing/synctest"
	"time"

	"kilabit.info/pakakeh.go/lib/test"
)

func TestPayload_Validate(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now().UTC().Unix()
		drifted1s := now + 1
		drifted2x := now + (DriftSeconds * 2)

		listCase := []struct {
			desc       string
			receiverID string
			expErr     string
			sender     Peer
			pload      Payload
		}{{
			desc: `ErrPayloadIssuer`,
			sender: Peer{
				ID: `me`,
			},
			pload: Payload{
				Issuer: `other`,
			},
			expErr: `payload: unknown issuer`,
		}, {
			desc: `ErrPayloadSubject`,
			sender: Peer{
				ID: `sender`,
				AllowedSubjects: map[string]struct{}{
					`create`: struct{}{},
					`update`: struct{}{},
				},
			},
			pload: Payload{
				Issuer:  `sender`,
				Subject: `delete`,
			},
			expErr: `payload: unknown subject: delete`,
		}, {
			desc: `ErrPayloadAudience`,
			sender: Peer{
				ID: `sender`,
				AllowedSubjects: map[string]struct{}{
					`create`: struct{}{},
					`update`: struct{}{},
				},
			},
			receiverID: `receiver`,
			pload: Payload{
				Issuer:   `sender`,
				Audience: `other`,
				Subject:  `create`,
			},
			expErr: `payload: invalid audience: other`,
		}, {
			desc: `ErrPayloadExpired`,
			pload: Payload{
				ExpiredAt: now,
			},
			expErr: `payload: expired: exp is 2000-01-01 00:00:00 +0000 UTC`,
		}, {
			desc: `NotBefore drifted 1s`,
			pload: Payload{
				NotBefore: drifted1s,
			},
		}, {
			desc: `ErrPayloadNotBefore`,
			pload: Payload{
				NotBefore: drifted2x,
			},
			expErr: `payload: token cannot be used yet: nbf is 2000-01-01 00:00:10 +0000 UTC`,
		}, {
			desc: `OK`,
			sender: Peer{
				ID: `sender`,
				AllowedSubjects: map[string]struct{}{
					`create`: struct{}{},
					`update`: struct{}{},
				},
			},
			receiverID: `receiver`,
			pload: Payload{
				Issuer:   `sender`,
				Audience: `receiver`,
				Subject:  `create`,
			},
		}}

		for _, tc := range listCase {
			err := tc.pload.Validate(tc.receiverID, tc.sender)
			if err != nil {
				test.Assert(t, tc.desc, tc.expErr, err.Error())
				continue
			}
		}
	})
}
