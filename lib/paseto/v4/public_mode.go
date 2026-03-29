// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
)

const publicHeader = `v4.public.`

// List of errors for [PublicMode.Unpack] and [PublicMode.Verify].
var (
	ErrSignature = errors.New(`invalid signature`)
)

// PublicMode contains ed25519 private and public key for signing and
// verifying message.
type PublicMode struct {
	peers map[string]paseto.Peer

	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey

	sync.Mutex
}

// NewPublicMode returns new instance of public mode from ed25519 seeds.
func NewPublicMode(seed []byte) (pmode *PublicMode) {
	pmode = &PublicMode{
		peers: make(map[string]paseto.Peer),
	}
	pmode.PrivateKey = ed25519.NewKeyFromSeed(seed)
	pmode.PublicKey = pmode.PrivateKey.Public().(ed25519.PublicKey)
	return pmode
}

// AddPeer adds the peer p to list of known peers for verifying incoming
// token.
// The only required fields in [paseto.Peer] is ID and Public.
func (pmode *PublicMode) AddPeer(peer paseto.Peer) (err error) {
	if len(peer.ID) == 0 {
		return errors.New(`empty peer ID`)
	}
	if len(peer.Public) == 0 {
		return errors.New(`empty public key`)
	}
	pmode.Lock()
	pmode.peers[peer.ID] = peer
	pmode.Unlock()
	return nil
}

// GetPeer returns the Peer and true if ID found in list of peers.
func (pmode *PublicMode) GetPeer(id string) (peer paseto.Peer, ok bool) {
	pmode.Lock()
	peer, ok = pmode.peers[id]
	pmode.Unlock()
	return peer, ok
}

// Pack returns the signed [paseto.Message] as public token.
// The token can be verified and decoded into Message again using
// [PublicMode.Unpack].
func (pmode *PublicMode) Pack(msg paseto.Message, implicit []byte) (
	token string, err error,
) {
	logp := `Pack`

	msg.RawPayload, err = json.Marshal(msg.Payload)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}
	msg.RawFooter, err = json.Marshal(msg.Footer)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}
	token, err = pmode.Sign(msg.RawPayload, msg.RawFooter, implicit)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}
	return token, nil
}

// RemovePeer removes the peer from list of peers based on its ID.
func (pmode *PublicMode) RemovePeer(id string) {
	pmode.Lock()
	delete(pmode.peers, id)
	pmode.Unlock()
}

// Sign returns the public token that has been signed with private key.
// The token contains header, payload, signature, and optional footer.
//
// The token can be verified to get the payload and footer again using
// [PublicMode.Verify].
func (pmode *PublicMode) Sign(payload, footer, implicit []byte) (
	token string, err error,
) {
	logp := `Sign`

	// Step 3: pack header, message, footer, and implicit.
	pae, err := paseto.PreAuthEncode([]byte(publicHeader),
		payload, footer, implicit)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	// Step 4: Sign pae.
	sig := ed25519.Sign(pmode.PrivateKey, pae)

	// Step 5: Pack all into token,
	var buf bytes.Buffer

	buf.WriteString(publicHeader)

	paysig := slices.Concat(payload, sig)
	n := base64.RawURLEncoding.EncodedLen(len(paysig))
	b64 := make([]byte, n)
	base64.RawURLEncoding.Encode(b64, paysig)
	buf.Write(b64)

	if len(footer) != 0 {
		buf.WriteByte('.')
		n = base64.RawURLEncoding.EncodedLen(len(footer))
		b64 = make([]byte, n)
		base64.RawURLEncoding.Encode(b64, footer)
		buf.Write(b64)
	}
	return buf.String(), nil
}

// Unpack verifies the token and returns it as [paseto.Message].
// The token is verified using public key from peer ID in the footer, that
// should has been added previously using [PublicMode.AddPeer].
//
// Note that, the [paseto.Payload] is not validated yet.
// It is up to the user to validate and handle the error.
func (pmode *PublicMode) Unpack(token string, implicit []byte, msg *paseto.Message) (
	err error,
) {
	logp := `Unpack`

	err = msg.Unpack(publicHeader, token, implicit)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = json.Unmarshal(msg.RawFooter, &msg.Footer)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	sender, ok := pmode.GetPeer(msg.Footer.PeerID)
	if !ok {
		return fmt.Errorf(`%s: unknown peer ID %s`, logp, msg.Footer.PeerID)
	}

	if !ed25519.Verify(sender.Public, msg.PAE, msg.Sig) {
		return fmt.Errorf(`%s: %w`, logp, ErrSignature)
	}

	err = json.Unmarshal(msg.RawPayload, &msg.Payload)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// Verify returns the payload and optional footer inside the token only if its
// signature is valid.
// The token is verified using the public key belong to PublicMode.
func (pmode *PublicMode) Verify(token string, implicit []byte) (
	payload, footer []byte, err error,
) {
	logp := `Verify`
	msg := &paseto.Message{}

	err = msg.Unpack(publicHeader, token, implicit)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	// Step 6: Verify the signature.
	if !ed25519.Verify(pmode.PublicKey, msg.PAE, msg.Sig) {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, ErrSignature)
	}

	return msg.RawPayload, msg.RawFooter, nil
}
