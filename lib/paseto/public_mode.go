// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var DefaultTTL = 60 * time.Second

//
// PublicMode implement the PASETO public mode to signing and verifying data
// using private key and one or more shared public keys.
//
type PublicMode struct {
	our   Key
	peers map[string]ed25519.PublicKey
}

//
// NewPublicMode create new PublicMode with our private key for signing
// outgoing token and list of peer public keys for verifying the incoming
// token.
//
func NewPublicMode(our Key, peers map[string]ed25519.PublicKey) (auth *PublicMode) {
	auth = &PublicMode{
		our:   our,
		peers: peers,
	}

	return auth
}

//
// Pack the data into token.
//
func (auth *PublicMode) Pack(data []byte, addFooter map[string]interface{}) (
	token string, err error,
) {
	now := time.Now()
	expiredAt := now.Add(DefaultTTL)
	jsonToken := JSONToken{
		Issuer:    auth.our.id,
		IssuedAt:  &now,
		ExpiredAt: &expiredAt,
		Data:      base64.StdEncoding.EncodeToString(data),
	}

	msg, err := json.Marshal(&jsonToken)
	if err != nil {
		return "", err
	}

	jsonFooter := JSONFooter{
		KID:  auth.our.id,
		Data: addFooter,
	}

	footer, err := json.Marshal(&jsonFooter)
	if err != nil {
		return "", err
	}

	return Sign(auth.our.private, msg, footer)
}

//
// Unpack the token to get the JSONToken and the data.
//
func (auth *PublicMode) Unpack(token string) (data []byte, addFooter map[string]interface{}, err error) {
	pieces := strings.Split(token, ".")
	if len(pieces) != 4 {
		return nil, nil, fmt.Errorf("invalid token format")
	}
	if pieces[0] != "v2" {
		return nil, nil, fmt.Errorf("unsupported protocol version " + pieces[0])
	}
	if pieces[1] != "public" {
		return nil, nil, fmt.Errorf("expecting public mode, got " + pieces[1])
	}

	footer, err := base64.RawURLEncoding.DecodeString(pieces[3])
	if err != nil {
		return nil, nil, err
	}

	jsonFooter := &JSONFooter{}
	err = json.Unmarshal(footer, jsonFooter)
	if err != nil {
		return nil, nil, err
	}
	peerKey, ok := auth.peers[jsonFooter.KID]
	if !ok {
		return nil, nil, fmt.Errorf("unknown peer key ID %s", jsonFooter.KID)
	}

	msgSig, err := base64.RawURLEncoding.DecodeString(pieces[2])
	if err != nil {
		return nil, nil, err
	}

	msg, err := Verify(peerKey, msgSig, footer)
	if err != nil {
		return nil, nil, err
	}

	jtoken := &JSONToken{}
	err = json.Unmarshal(msg, jtoken)
	if err != nil {
		return nil, nil, err
	}

	err = jtoken.Validate()
	if err != nil {
		return nil, nil, err
	}

	data, err = base64.StdEncoding.DecodeString(jtoken.Data)
	if err != nil {
		return nil, nil, err
	}

	return data, jsonFooter.Data, nil
}
