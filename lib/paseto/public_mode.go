// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package paseto

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	headerAuthorization  = "Authorization"
	paramNameAccessToken = "access_token"
	keyBearer            = "bearer"
)

// DefaultTTL define the time-to-live of each token, by setting ExpiredAt to
// current time + DefaultTTL.
// If you want longer token, increase this value before using Pack().
var DefaultTTL = 60 * time.Second

// PublicMode implement the PASETO public mode to signing and verifying data
// using private key and one or more shared public keys.
// The PublicMode contains list of peer public keys for verifying the incoming
// token.
type PublicMode struct {
	peers *keys
	our   Key
}

// NewPublicMode create new PublicMode with our private key for signing
// outgoing token.
func NewPublicMode(our Key) (auth *PublicMode, err error) {
	if len(our.ID) == 0 {
		return nil, errors.New(`empty key ID`)
	}
	if len(our.Private) == 0 {
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
	publicToken *PublicToken, err error,
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

// AddPeer add a key to list of known peers for verifying incoming token.
// The only required fields in Key is ID and Public.
func (auth *PublicMode) AddPeer(k Key) (err error) {
	if len(k.ID) == 0 {
		return errors.New(`empty key ID`)
	}
	if len(k.Public) == 0 {
		return errors.New(`empty public key`)
	}
	auth.peers.upsert(k)
	return nil
}

// GetPeerKey get the peer's key based on key ID.
func (auth *PublicMode) GetPeerKey(id string) (k Key, ok bool) {
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
	now := time.Now().Round(time.Second)
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

// Unpack the token to get the JSONToken and the data.
func (auth *PublicMode) Unpack(token string) (publicToken *PublicToken, err error) {
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

	publicToken = &PublicToken{}

	rawfooter, err := base64.RawURLEncoding.DecodeString(pieces[3])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawfooter, &publicToken.Footer)
	if err != nil {
		return nil, err
	}
	peerKey, ok := auth.peers.get(publicToken.Footer.KID)
	if !ok {
		return nil, fmt.Errorf("unknown peer key ID %s",
			publicToken.Footer.KID)
	}

	msgSig, err := base64.RawURLEncoding.DecodeString(pieces[2])
	if err != nil {
		return nil, err
	}

	msg, err := Verify(peerKey.Public, msgSig, rawfooter)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(msg, &publicToken.Token)
	if err != nil {
		return nil, err
	}

	err = publicToken.Token.Validate(auth.our.ID, peerKey)
	if err != nil {
		return nil, err
	}

	publicToken.Data, err = base64.StdEncoding.DecodeString(publicToken.Token.Data)
	if err != nil {
		return nil, err
	}

	return publicToken, nil
}
