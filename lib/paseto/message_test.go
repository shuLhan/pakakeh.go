// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package paseto

import (
	"encoding/json"
	"testing"
	"testing/synctest"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewMessage(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sender := Peer{
			ID: `sender`,
		}
		receiver := Peer{
			ID: `receiver`,
		}
		msg := NewMessage(sender, receiver, `test`)

		gotmsg, err := json.MarshalIndent(msg, ``, `  `)
		if err != nil {
			t.Fatal(err)
		}

		expmsg := `{
  "Footer": {
    "peer_id": "sender"
  },
  "Payload": {
    "iss": "sender",
    "sub": "test",
    "aud": "receiver",
    "exp": 946684860,
    "nbf": 946684800,
    "iat": 946684800
  }
}`
		test.Assert(t, `NewMessage`, expmsg, string(gotmsg))
	})
}
