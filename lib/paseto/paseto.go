// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package paseto provide a simple, ready to use, opinionated implementation
// of Platform-Agnostic SEcurity TOkens (PASETOs) v2 as defined in draft of
// RFC 01 [1].
//
// Limitation
//
// This implementation only support PASETO Protocol v2.
//
// Local mode
//
// The local mode use crypto/rand package to generate random nonce and hashed
// with blake2b.
//
// Public mode
//
// The public mode focus on signing and verifing data, everything else is
// handled and filled automatically.
//
// Steps for sender when generating new token, the Pack() method,
//
//	* Prepare the JSON token claims, set
//	** Issuer "iss" to PublicMode.our.ID
//	** Subject "sub" to subject value from parameter
//	** Audience "aud" to audience value from parameter
//	** IssuedAt to current time
//	** NotBefore to current time
//	** ExpiredAt to current time + 60 seconds
//	** Data field to base64 encoded of data value from parameter
//	* Prepare the JSON footer, set
//	** Key ID "kid" to PublicMode.our.ID
//
// The user's claims data is stored using key "data" inside the JSON token,
// encoded using base64 (with padding).
// Additional footer data can be added on the Data field.
//
// Overall, the following JSONToken and JSONFooter is generated for each
// token,
//
//	JSONToken:{
//		"iss": <Key.ID>,
//		"sub": <Subject parameter>,
//		"aud": <Audience parameter>
//		"exp": <time.Now() + TTL>,
//		"iat": <time.Now()>,
//		"nbf": <time.Now()>,
//		"data": <base64.StdEncoding.EncodeToString(userData)>,
//	}
//	JSONFooter:{
//		"kid": <Key.ID>,
//		"data": {}
//	}
//
// On the receiver side, they will have list of registered peers Key (include
// ID, public Key, and list of allowed subject).
//
//	PublicMode:{
//		peers: map[Key.ID]Key{
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
// Step for receiver to process the token, the Unpack() method,
//
//	* Decode the token footer
//	* Get the registered public key based on "kid" value in token footer
//	** If no peers key exist matched with "kid" value, reject the token
//	* Verify the token using the peer public key
//	** If verification failed, reject the token
//	* Validate the token
//	** The Issuer must equal to peer ID
//	** The Audience must equal to receiver ID
//	** If the peer AllowedSubjects is not empty, the Subject must be in
//	one of them
//	** The current time must be after IssuedAt
//	** The current time must be after NotBefore
//	** The current time must be before ExpiredAt
//	** If one of the above condition is not passed, it will return an error.
//
// References
//
// [1] https://github.com/paragonie/paseto/blob/master/docs/RFC/draft-paragon-paseto-rfc-01.txt
//
package paseto

import (
	"bytes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
)

const (
	randNonceSize = 24
)

var (
	headerModePublic = []byte("v2.public.")
	headerModeLocal  = []byte("v2.local.")
)

//
// Encrypt given the shared key, encrypt the plain message and generate the
// "local" token with optional footer.
//
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

	pieces := [][]byte{headerModeLocal, nonce, footer}

	m2, err := pae(pieces)
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
		return "", nil
	}

	if len(footer) > 0 {
		buf.WriteByte('.')

		n = base64.RawURLEncoding.EncodedLen(len(footer))
		dst = make([]byte, n)
		base64.RawURLEncoding.Encode(dst, footer)
		_, err = buf.Write(dst)
		if err != nil {
			return "", nil
		}
	}

	return buf.String(), nil
}

//
// Decrypt given a shared key and encrypted token, decrypt the token to get
// the message.
//
func Decrypt(aead cipher.AEAD, token string) (plain, footer []byte, err error) {
	pieces := strings.Split(token, ".")
	if len(pieces) < 3 || len(pieces) > 4 {
		return nil, nil, errors.New("invalid token format")
	}
	if pieces[0] != "v2" {
		return nil, nil, fmt.Errorf("unsupported protocol version " + pieces[0])
	}
	if pieces[1] != "local" {
		return nil, nil, fmt.Errorf("expecting local mode, got " + pieces[1])
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

	m2, err := pae([][]byte{headerModeLocal, nonce, footer})
	if err != nil {
		return nil, nil, err
	}

	plain, err = aead.Open(nil, nonce, cipher, m2)
	if err != nil {
		return nil, nil, err
	}

	return plain, footer, nil
}

//
// Sign given an Ed25519 secret key "sk", a message "m", and optional footer
// "f" (which defaults to empty string); sign the message "m" and generate the
// public token.
//
func Sign(sk ed25519.PrivateKey, m, f []byte) (token string, err error) {
	pieces := [][]byte{headerModePublic, m, f}

	m2, err := pae(pieces)
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

//
// Verify given a public key "pk", a signed message "sm" (that has been
// decoded from base64), and optional footer "f" (also that has been decoded
// from base64 string); verify that the signature is valid for the message.
//
func Verify(pk ed25519.PublicKey, sm, f []byte) (msg []byte, err error) {
	if len(sm) <= 64 {
		return nil, fmt.Errorf("invalid signed message length")
	}

	msg = sm[:len(sm)-64]
	sig := sm[len(sm)-64:]
	pieces := [][]byte{headerModePublic, msg, f}

	msg2, err := pae(pieces)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(pk, msg2, sig) {
		return nil, fmt.Errorf("invalid message signature")
	}

	return msg, nil
}

func pae(pieces [][]byte) (b []byte, err error) {
	var buf bytes.Buffer

	b, err = le64(int64(len(pieces)))
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}

	for x := 0; x < len(pieces); x++ {
		b, err = le64(int64(len(pieces[x])))
		if err != nil {
			return nil, err
		}

		_, err = buf.Write(b)
		if err != nil {
			return nil, err
		}

		_, err = buf.Write(pieces[x])
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func le64(n int64) (out []byte, err error) {
	var buf bytes.Buffer

	for x := 0; x < 8; x++ {
		if x == 7 {
			n &= 127
		}
		_, err = buf.WriteRune(rune(n & 255))
		if err != nil {
			return out, err
		}
		n = n >> 8
	}
	return buf.Bytes(), nil
}
