// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

package pasetov2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kilabit.info/pakakeh.go/lib/paseto"
)

const (
	headerAuthorization  = "Authorization"
	paramNameAccessToken = "access_token"
	keyBearer            = "bearer"
)

// PublicMode implement the PASETO public mode to signing and verifying data
// using private key and one or more shared public keys.
// The PublicMode contains list of peer public keys for verifying the incoming
// token.
type PublicMode struct {
	peers *keys
	our   paseto.Peer
}

// NewPublicMode create new PublicMode with our private key for signing
// outgoing token.
func NewPublicMode(our paseto.Peer) (auth *PublicMode, err error) {
	if len(our.ID) == 0 {
		return nil, errors.New(`empty key ID`)
	}
	if len(our.PrivateKey) == 0 {
		return nil, errors.New(`empty private key`)
	}
	auth = &PublicMode{
		our:   our,
		peers: newKeys(),
	}
	return auth, nil
}

// UnpackHTTPRequest unpack token from HTTP request header "Authorization" or
// from query parameter "access_token".
func (auth *PublicMode) UnpackHTTPRequest(req *http.Request) (
	msg *paseto.Message, err error,
) {
	if req == nil {
		return nil, errors.New(`empty HTTP request`)
	}

	var token string

	headerAuth := req.Header.Get(headerAuthorization)
	if len(headerAuth) == 0 {
		token = req.Form.Get(paramNameAccessToken)
		if len(token) == 0 {
			return nil, errors.New(`missing access token`)
		}
	} else {
		vals := strings.Fields(headerAuth)
		if len(vals) != 2 {
			return nil, fmt.Errorf("invalid Authorization: %s", headerAuth)
		}
		if strings.ToLower(vals[0]) != keyBearer {
			return nil, fmt.Errorf("invalid Authorization: expecting %q, got %q",
				keyBearer, vals[0])
		}
		token = vals[1]
	}

	return auth.Unpack(token)
}

// AddPeer add a peer to list of known peers for verifying incoming token.
// The only required fields in [paseto.Peer] is ID and Public.
func (auth *PublicMode) AddPeer(k paseto.Peer) (err error) {
	if len(k.ID) == 0 {
		return errors.New(`empty key ID`)
	}
	if len(k.PublicKey) == 0 {
		return errors.New(`empty public key`)
	}
	auth.peers.upsert(k)
	return nil
}

// GetPeerKey get the peer's key based on key ID.
func (auth *PublicMode) GetPeerKey(id string) (k paseto.Peer, ok bool) {
	return auth.peers.get(id)
}

// RemovePeer remove peer's key from list.
func (auth *PublicMode) RemovePeer(id string) {
	auth.peers.delete(id)
}

// Pack the data into token.
func (auth *PublicMode) Pack(audience, subject string, data []byte, footer map[string]any) (
	token string, err error,
) {
	now := time.Now().UTC().Unix()
	expiredAt := now + paseto.DefaultTTL
	jsonToken := paseto.Payload{
		Issuer:    auth.our.ID,
		Subject:   subject,
		Audience:  audience,
		IssuedAt:  now,
		NotBefore: now,
		ExpiredAt: expiredAt,
		Data:      base64.StdEncoding.EncodeToString(data),
	}

	msg, err := json.Marshal(&jsonToken)
	if err != nil {
		return "", err
	}

	jsonFooter := paseto.Footer{
		PeerID: auth.our.ID,
		Data:   footer,
	}

	rawfooter, err := json.Marshal(&jsonFooter)
	if err != nil {
		return "", err
	}

	return Sign(auth.our.PrivateKey, msg, rawfooter)
}

// Unpack the token to get the Payload and the data.
func (auth *PublicMode) Unpack(token string) (msg *paseto.Message, err error) {
	pieces := strings.Split(token, ".")
	if len(pieces) != 4 {
		return nil, errors.New(`invalid token format`)
	}
	if pieces[0] != "v2" {
		return nil, errors.New(`unsupported protocol version ` + pieces[0])
	}
	if pieces[1] != "public" {
		return nil, errors.New(`expecting public mode, got ` + pieces[1])
	}

	msg = &paseto.Message{}

	rawfooter, err := base64.RawURLEncoding.DecodeString(pieces[3])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawfooter, &msg.Footer)
	if err != nil {
		return nil, err
	}
	peerKey, ok := auth.peers.get(msg.Footer.PeerID)
	if !ok {
		return nil, fmt.Errorf("unknown peer key ID %s",
			msg.Footer.PeerID)
	}

	msgSig, err := base64.RawURLEncoding.DecodeString(pieces[2])
	if err != nil {
		return nil, err
	}

	payload, err := Verify(peerKey.PublicKey, msgSig, rawfooter)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(payload, &msg.Payload)
	if err != nil {
		return nil, err
	}

	err = msg.Payload.Validate(auth.our.ID, peerKey)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(msg.Payload.Data.(string))
	if err != nil {
		return nil, err
	}
	msg.Payload.Data = data

	return msg, nil
}
