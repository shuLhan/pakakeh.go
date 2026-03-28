// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"
)

const localHeader = `v4.local.`
const localKeyPrefixEncryption = `paseto-encryption-key`
const localKeyPrefixAuth = `paseto-auth-key-for-aead`

// LocalMode contains the secret for Paseto v4 local protocol.
type LocalMode struct {
	v [32]byte
}

// NewLocalMode returns new local key.
func NewLocalMode(key [32]byte) (lmode *LocalMode) {
	lmode = &LocalMode{}
	copy(lmode.v[:], key[:])
	return lmode
}

// Encrypt encrypts the plain text with optional footer and implicit using
// LocalMode as defined in [paseto-v4-encrypt].
//
// [paseto-v4-encrypt]: https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md#encrypt
func (lmode LocalMode) Encrypt(plain, footer, implicit []byte) (token string, err error) {
	logp := `Encrypt`

	var salt [32]byte
	rand.Read(salt[:])

	token, err = lmode.encrypt(salt, plain, footer, implicit)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}
	return token, nil
}

func (lmode LocalMode) encrypt(salt [32]byte, plain, footer, implicit []byte) (string, error) {
	// Step 4: generate encryption key, authentication key, and nonce.
	encKey, authKey, nonce, err := lmode.gen(salt)
	if err != nil {
		return ``, err
	}

	// Step 5: encrpyt the plain text.
	cipher, err := chacha20.NewUnauthenticatedCipher(encKey, nonce)
	if err != nil {
		return ``, err
	}

	ciphertext := make([]byte, len(plain))
	cipher.XORKeyStream(ciphertext, plain)

	// Step 6: generate pre-authentication encoding.
	pae, err := paseto.PreAuthEncode([]byte(localHeader), salt[:], ciphertext, footer, implicit)
	if err != nil {
		return ``, err
	}

	// Step 7: MAC the pae.
	hash, err := blake2b.New256(authKey)
	if err != nil {
		return ``, err
	}
	hash.Write(pae)
	mac := hash.Sum(nil)

	// Step 8: create the paseto token from cipher text and mac.
	var buf bytes.Buffer

	buf.WriteString(localHeader)

	payload := slices.Concat(salt[:], ciphertext, mac)
	n := base64.RawURLEncoding.EncodedLen(len(payload))
	b64 := make([]byte, n)
	base64.RawURLEncoding.Encode(b64, payload)
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

// Decrypt returns the plain text of encrypted message inside token with
// optional footer.
func (lmode LocalMode) Decrypt(token string, implicit []byte) (plain, footer []byte, err error) {
	logp := `Decrypt`

	// Step 3: verify the header.
	token, found := strings.CutPrefix(token, localHeader)
	if !found {
		return nil, nil, fmt.Errorf(`%s: invalid header, want %s`, logp, localHeader)
	}

	token, footerb64, found := strings.Cut(token, `.`)
	if found {
		footer, err = base64.RawURLEncoding.DecodeString(footerb64)
		if err != nil {
			return nil, nil, fmt.Errorf(`%s: invalid footer: %w`, logp, err)
		}
	}

	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	// Step 4: decode the
	if len(payload) <= 64 { // 64 is size of nonce + mac.
		return nil, nil, fmt.Errorf(`%s: invalid payload length`, logp)
	}

	var salt [32]byte
	copy(salt[:], payload[:32])
	mac := payload[len(payload)-32:]
	ciphertext := payload[32 : len(payload)-32]

	// Step 5: generate encryption key, authentication key, and nonce.
	encKey, authKey, nonce, err := lmode.gen(salt)
	if err != nil {
		return nil, nil, err
	}

	// Step 6: generate pre-authentication encoding.
	pae, err := paseto.PreAuthEncode([]byte(localHeader), salt[:], ciphertext, footer, implicit)
	if err != nil {
		return nil, nil, err
	}

	// Step 7: MAC the pae.
	hash, err := blake2b.New256(authKey)
	if err != nil {
		return nil, nil, err
	}
	hash.Write(pae)
	gotmac := hash.Sum(nil)

	// Step 8: Compare MAC.
	if !bytes.Equal(mac, gotmac) {
		return nil, nil, fmt.Errorf(`%s: MAC mismatch`, logp)
	}

	// Step 9: Decrypt the cipher text.
	cipher, err := chacha20.NewUnauthenticatedCipher(encKey, nonce)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	plain = make([]byte, len(ciphertext))
	cipher.XORKeyStream(plain, ciphertext)

	return plain, footer, nil
}

// gen generates new encryption key, authentication key, and nonce.
func (lmode LocalMode) gen(salt [32]byte) (encKey, authKey, nonce []byte, err error) {
	logp := `gen`

	hash, err := blake2b.New(56, lmode.v[:])
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	hash.Write([]byte(localKeyPrefixEncryption))
	hash.Write(salt[:])
	out := hash.Sum(nil)
	encKey = out[:32]
	nonce = out[32:56]

	// Create new hash with 32 bytes size.
	hash, err = blake2b.New256(lmode.v[:])
	if err != nil {
		return nil, nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	hash.Write([]byte(localKeyPrefixAuth))
	hash.Write(salt[:])
	authKey = hash.Sum(nil)

	return encKey, authKey, nonce, nil
}

// Hex returns the key as hexadecimal string.
func (lmode LocalMode) Hex() string {
	return hex.EncodeToString(lmode.v[:])
}
