// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestPayload_Validate(t *testing.T) {
	now := time.Now().Round(time.Second)
	peer := Peer{}

	issued1sAgo := now.Add(-1 * time.Second)
	issued6sAgo := now.Add(-6 * time.Second)

	listCase := []struct {
		desc   string
		pload  *Payload
		expErr string
	}{{
		desc: `With IssuedAt less than current time`,
		pload: &Payload{
			IssuedAt: &issued1sAgo,
		},
	}, {
		desc: `With IssuedAt greater than drift`,
		pload: &Payload{
			IssuedAt: &issued6sAgo,
		},
		expErr: `payload: issued-at is after current time`,
	}}

	for _, tc := range listCase {
		err := tc.pload.Validate(``, peer)
		if err != nil {
			test.Assert(t, tc.desc, tc.expErr, err.Error())
			continue
		}
	}
}
