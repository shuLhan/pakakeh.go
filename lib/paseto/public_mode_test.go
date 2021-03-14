// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestPublicMode_UnpackHTTPRequest(t *testing.T) {
	subjectMessage := "message"

	senderSK, _ := hex.DecodeString("e9ae9c7eae2fce6fd6727b5ca8df0fbc0aa60a5ffb354d4fdee1729e4e1463688d2160a4dc71a9a697d6ad6424da3f9dd18a259cdd51b0ae2b521e998b82d36e")
	senderPK, _ := hex.DecodeString("8d2160a4dc71a9a697d6ad6424da3f9dd18a259cdd51b0ae2b521e998b82d36e")
	ourKey := Key{
		ID:      "sender",
		Private: ed25519.PrivateKey(senderSK),
		Public:  ed25519.PublicKey(senderPK),
		AllowedSubjects: map[string]struct{}{
			subjectMessage: struct{}{},
		},
	}

	auth, err := NewPublicMode(ourKey)
	if err != nil {
		t.Fatal(err)
	}

	err = auth.AddPeer(ourKey)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("This is a signed message")

	token, err := auth.Pack(ourKey.ID, subjectMessage, data, nil)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc     string
		req      *http.Request
		expData  []byte
		expError string
	}{{
		desc:     "With request is nil",
		expError: "empty HTTP request",
	}, {
		desc:     "With no Authorization header",
		req:      &http.Request{},
		expError: "missing access token",
	}, {
		desc: "With invalid Authorization header",
		req: &http.Request{
			Header: map[string][]string{
				headerAuthorization: []string{
					"Beer " + token,
				},
			},
		},
		expError: `invalid Authorization: expecting "bearer", got "Beer"`,
	}, {
		desc: "With valid token in header",
		req: &http.Request{
			Header: map[string][]string{
				headerAuthorization: []string{
					"Bearer " + token,
				},
			},
		},
		expData: data,
	}, {
		desc: "With valid token in query parameter",
		req: &http.Request{
			Form: map[string][]string{
				paramNameAccessToken: []string{
					token,
				},
			},
		},
		expData: data,
	}}

	for _, c := range cases {
		got, err := auth.UnpackHTTPRequest(c.req)
		if err != nil {
			test.Assert(t, c.desc, c.expError, err.Error())
			continue
		}

		test.Assert(t, c.desc, c.expData, got.Data)
	}
}
