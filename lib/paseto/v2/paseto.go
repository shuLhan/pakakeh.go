// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

// Package paseto provide a simple, ready to use, opinionated implementation
// of Platform-Agnostic SEcurity TOkens (PASETOs) v2 as defined in
// [paseto-rfc-01].
//
// Deprecated: This package has been replaced with new standard, paseto v4.
// See git.sr.ht/~shulhan/pakakeh.go/lib/paseto/v4.
//
// See the examples below for quick reference.
//
// # Limitation
//
// This implementation only support PASETO Protocol v2.
//
// # Local mode
//
// The local mode use crypto/rand package to generate random nonce and hashed
// with blake2b.
//
// # Public mode
//
// The public mode focus on signing and verifing data, everything else is
// handled and filled automatically.
//
// # Sender side
//
// Steps for generating new token, the [PublicMode.Pack] method, on sender
// side:
//
// (1) Prepare the JSON token claims, set
//
//   - Issuer "iss" to PublicMode.our.ID
//   - Subject "sub" to subject value from parameter
//   - Audience "aud" to audience value from parameter
//   - IssuedAt to current time
//   - NotBefore to current time
//   - ExpiredAt to current time + 60 seconds
//   - Data field to base64 encoded of data value from parameter
//
// (2) Prepare the JSON footer, set
//
//   - Peer ID "peer_id" to PublicMode.our.ID
//
// The user's claims data is stored using key "data" inside the JSON token,
// encoded using base64 (with padding).
// Additional footer data can be added on the Data field.
//
// Overall, the following [paseto.Payload] and [paseto.Footer] is generated
// for each token,
//
//	Payload:{
//		"iss": <Peer.ID>,
//		"sub": <Subject parameter>,
//		"aud": <Audience parameter>
//		"exp": <time.Now() + TTL>,
//		"iat": <time.Now()>,
//		"nbf": <time.Now()>,
//		"data": <base64.StdEncoding.EncodeToString(userData)>,
//	}
//	Footer:{
//		"peer_id": <Peer.ID>,
//		"data": {}
//	}
//
// # Receiver side
//
// On the receiver side, they will have list of registered peers (including
// their ID, public key, and list of allowed subject).
//
//	PublicMode:{
//		peers: map[Peer.ID]Peer{
//			Public: <ed25519.PublicKey>,
//			AllowedSubjects: map[string]struct{}{
//				"/api/x": struct{}{},
//				"/api/y:read": struct{}{},
//				"/api/z:write": struct{}{},
//				...
//			},
//		},
//	}
//
// Step for receiver to process the token, the [PublicMode.Unpack] method,
//
// (1) Decode the token footer.
//
// (2) Get the registered public key based on "peer_id" value in token footer.
// If no peers key exist matched with "peer_id" value, reject the token.
//
// (3) Verifies the token using the peer public key.
// If verification failed, reject the token.
//
// (4) Validates the token.
//
// [paseto-rfc-01]: https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version2.md
package pasetov2

import (
	"bytes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
	"golang.org/x/crypto/blake2b"
)

const (
	randNonceSize = 24
)

var (
	headerModePublic = []byte("v2.public.")
	headerModeLocal  = []byte("v2.local.")
)

// Encrypt given the shared key, encrypt the plain message and generate the
// "local" token with optional footer.
func Encrypt(aead cipher.AEAD, plain, footer []byte) (token string, err error) {
	nonce := make([]byte, randNonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	return encrypt(aead, nonce, plain, footer)
}

func encrypt(aead cipher.AEAD, nonce, plain, footer []byte) (token string, err error) {
	b2b, err := blake2b.New(randNonceSize, nonce)
	if err != nil {
		return "", err
	}

	_, err = b2b.Write(plain)
	if err != nil {
		return "", err
	}

	nonce = b2b.Sum(nil)

	m2, err := paseto.PreAuthEncode(headerModeLocal, nonce, footer)
	if err != nil {
		return "", err
	}

	cipher := aead.Seal(nil, nonce, plain, m2)

	var buf bytes.Buffer

	_, err = buf.Write(headerModeLocal)
	if err != nil {
		return "", err
	}

	sc := make([]byte, 0, len(nonce)+len(cipher))
	sc = append(sc, nonce...)
	sc = append(sc, cipher...)

	n := base64.RawURLEncoding.EncodedLen(len(sc))
	dst := make([]byte, n)
	base64.RawURLEncoding.Encode(dst, sc)
	_, err = buf.Write(dst)
	if err != nil {
		return ``, err
	}

	if len(footer) > 0 {
		buf.WriteByte('.')

		n = base64.RawURLEncoding.EncodedLen(len(footer))
		dst = make([]byte, n)
		base64.RawURLEncoding.Encode(dst, footer)
		_, err = buf.Write(dst)
		if err != nil {
			return ``, err
		}
	}

	return buf.String(), nil
}

// Decrypt given a shared key and encrypted token, decrypt the token to get
// the message.
func Decrypt(aead cipher.AEAD, token string) (plain, footer []byte, err error) {
	pieces := strings.Split(token, ".")
	if len(pieces) < 3 || len(pieces) > 4 {
		return nil, nil, errors.New("invalid token format")
	}
	if pieces[0] != "v2" {
		return nil, nil, errors.New(`unsupported protocol version ` + pieces[0])
	}
	if pieces[1] != "local" {
		return nil, nil, errors.New(`expecting local mode, got ` + pieces[1])
	}

	if len(pieces) == 4 {
		footer, err = base64.RawURLEncoding.DecodeString(pieces[3])
		if err != nil {
			return nil, nil, err
		}
	}

	src, err := base64.RawURLEncoding.DecodeString(pieces[2])
	if err != nil {
		return nil, nil, err
	}

	nonce := src[:randNonceSize]
	cipher := src[randNonceSize:]

	if len(cipher) < aead.NonceSize() {
		return nil, nil, errors.New("ciphertext too short")
	}

	m2, err := paseto.PreAuthEncode(headerModeLocal, nonce, footer)
	if err != nil {
		return nil, nil, err
	}

	plain, err = aead.Open(nil, nonce, cipher, m2)
	if err != nil {
		return nil, nil, err
	}

	return plain, footer, nil
}

// Sign given an Ed25519 secret key "sk", a message "m", and optional footer
// "f" (which defaults to empty string); sign the message "m" and generate the
// public token.
func Sign(sk ed25519.PrivateKey, m, f []byte) (token string, err error) {
	m2, err := paseto.PreAuthEncode(headerModePublic, m, f)
	if err != nil {
		return "", err
	}

	sig := ed25519.Sign(sk, m2)

	var buf bytes.Buffer

	_, err = buf.Write(headerModePublic)
	if err != nil {
		return "", err
	}

	sm := make([]byte, 0, len(m)+len(sig))
	sm = append(sm, m...)
	sm = append(sm, sig...)

	n := base64.RawURLEncoding.EncodedLen(len(sm))
	dst := make([]byte, n)
	base64.RawURLEncoding.Encode(dst, sm)

	_, err = buf.Write(dst)
	if err != nil {
		return "", err
	}

	if len(f) > 0 {
		_ = buf.WriteByte('.')

		n = base64.RawURLEncoding.EncodedLen(len(f))
		dst = make([]byte, n)
		base64.RawURLEncoding.Encode(dst, f)

		_, err = buf.Write(dst)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// Verify given a public key "pk", a signed message "sm" (that has been
// decoded from base64), and optional footer "f" (also that has been decoded
// from base64 string); verify that the signature is valid for the message.
func Verify(pk ed25519.PublicKey, sm, f []byte) (msg []byte, err error) {
	if len(sm) <= 64 {
		return nil, errors.New(`invalid signed message length`)
	}

	msg = sm[:len(sm)-64]
	sig := sm[len(sm)-64:]

	msg2, err := paseto.PreAuthEncode(headerModePublic, msg, f)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(pk, msg2, sig) {
		return nil, errors.New(`invalid message signature`)
	}

	return msg, nil
}
