// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//
// DefaultTTL define the time-to-live of each token, by setting ExpiredAt to
// current time + DefaultTTL.
// If you want longer token, increase this value before using Pack().
//
var DefaultTTL = 60 * time.Second

//
// PublicMode implement the PASETO public mode to signing and verifying data
// using private key and one or more shared public keys.
// The PublicMode contains list of peer public keys for verifying the incoming
// token.
//
type PublicMode struct {
	our   Key
	peers *keys
}

//
// NewPublicMode create new PublicMode with our private key for signing
// outgoing token.
//
func NewPublicMode(our Key) (auth *PublicMode) {
	auth = &PublicMode{
		our:   our,
		peers: newKeys(),
	}
	return auth
}

//
// AddPeer add a key to list of known peers for verifying incoming token.
// The Key.Public
//
func (auth *PublicMode) AddPeer(k Key) (err error) {
	if len(k.ID) == 0 {
		return fmt.Errorf("empty key ID")
	}
	if len(k.Public) == 0 {
		return fmt.Errorf("empty public key")
	}
	auth.peers.upsert(k)
	return nil
}

//
// RemovePeer remove peer's key from list.
//
func (auth *PublicMode) RemovePeer(id string) {
	auth.peers.delete(id)
}

//
// Pack the data into token.
//
func (auth *PublicMode) Pack(audience, subject string, data []byte, footer map[string]interface{}) (
	token string, err error,
) {
	now := time.Now()
	expiredAt := now.Add(DefaultTTL)
	jsonToken := JSONToken{
		Issuer:    auth.our.ID,
		Subject:   subject,
		Audience:  audience,
		IssuedAt:  &now,
		NotBefore: &now,
		ExpiredAt: &expiredAt,
		Data:      base64.StdEncoding.EncodeToString(data),
	}

	msg, err := json.Marshal(&jsonToken)
	if err != nil {
		return "", err
	}

	jsonFooter := JSONFooter{
		KID:  auth.our.ID,
		Data: footer,
	}

	rawfooter, err := json.Marshal(&jsonFooter)
	if err != nil {
		return "", err
	}

	return Sign(auth.our.Private, msg, rawfooter)
}

//
// Unpack the token to get the JSONToken and the data.
//
func (auth *PublicMode) Unpack(token string) (data []byte, footer map[string]interface{}, err error) {
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

	rawfooter, err := base64.RawURLEncoding.DecodeString(pieces[3])
	if err != nil {
		return nil, nil, err
	}

	jsonFooter := &JSONFooter{}
	err = json.Unmarshal(rawfooter, jsonFooter)
	if err != nil {
		return nil, nil, err
	}
	peerKey, ok := auth.peers.get(jsonFooter.KID)
	if !ok {
		return nil, nil, fmt.Errorf("unknown peer key ID %s", jsonFooter.KID)
	}

	msgSig, err := base64.RawURLEncoding.DecodeString(pieces[2])
	if err != nil {
		return nil, nil, err
	}

	msg, err := Verify(peerKey.Public, msgSig, rawfooter)
	if err != nil {
		return nil, nil, err
	}

	jtoken := &JSONToken{}
	err = json.Unmarshal(msg, jtoken)
	if err != nil {
		return nil, nil, err
	}

	err = jtoken.Validate(auth.our.ID, peerKey)
	if err != nil {
		return nil, nil, err
	}

	data, err = base64.StdEncoding.DecodeString(jtoken.Data)
	if err != nil {
		return nil, nil, err
	}

	return data, jsonFooter.Data, nil
}
