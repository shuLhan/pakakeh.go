// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
)

const publicHeader = `v4.public.`

// PublicMode contains ed25519 private and public key for signing and
// verifying message.
type PublicMode struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

// NewPublicMode returns new instance of public mode from ed25519 seeds.
func NewPublicMode(seed []byte) (pmode *PublicMode) {
	pmode = &PublicMode{}
	pmode.priv = ed25519.NewKeyFromSeed(seed)
	pmode.pub = pmode.priv.Public().(ed25519.PublicKey)
	return pmode
}

// Sign returns the public token that has been signing with private key.
// The token contains msg and optional footer.
func (pmode *PublicMode) Sign(msg, footer, implicit []byte) (token string, err error) {
	logp := `Sign`

	// Step 3: pack header, message, footer, and implicit.
	pae, err := paseto.PreAuthEncode([]byte(publicHeader), msg, footer, implicit)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	// Step 4: Sign pae.
	sig := ed25519.Sign(pmode.priv, pae)

	// Step 5: Pack all into token,
	var buf bytes.Buffer

	buf.WriteString(publicHeader)

	payload := slices.Concat(msg, sig)
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

// Verify returns the msg, with optional footer data, inside the token only if
// its signature is valid.
func (pmode *PublicMode) Verify(token string, implicit []byte) (msg, footer []byte, err error) {
	logp := `Verify`

	// Step 3: verify the header and unpack the footer if it exists.
	token, found := strings.CutPrefix(token, publicHeader)
	if !found {
		return nil, nil, fmt.Errorf(`%s: invalid header, want %s`, logp, publicHeader)
	}
	token, footerb64, found := strings.Cut(token, `.`)
	if found {
		footer, err = base64.RawURLEncoding.DecodeString(footerb64)
		if err != nil {
			return nil, nil, fmt.Errorf(`%s: invalid footer: %w`, logp, err)
		}
	}

	// Step 4: Decodes the payload.
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	lenpayload := len(payload)
	if lenpayload <= ed25519.SignatureSize {
		return nil, nil, fmt.Errorf(`%s: invalid payload size, want %d got %d`,
			logp, ed25519.SignatureSize, len(payload))
	}
	msg = payload[:lenpayload-64]
	sig := payload[lenpayload-64:]

	// Step 5: Generate PAE.
	pae, err := paseto.PreAuthEncode([]byte(publicHeader), msg, footer, implicit)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	// Step 6: Verify the signature.
	if !ed25519.Verify(pmode.pub, pae, sig) {
		return nil, nil, fmt.Errorf(`%s: invalid message signature`, logp)
	}

	return msg, footer, nil
}
