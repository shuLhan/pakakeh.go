// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// Message defines the payload be signed and verified by sender/receiver.
type Message struct {
	// The following fields are filled after Unpack and MUST not be used
	// directly.
	RawPayload []byte
	RawFooter  []byte
	Sig        []byte
	PAE        []byte

	// The Payload and optional Footer to be packed and signed into token.
	Footer  Footer
	Payload Payload
}

// NewMessage returns [Message] where all fields except [Payload.Data] and
// [Footer.Data] are set using send, receiver, subject, and current time.
func NewMessage(sender, receiver Peer, subject string) (msg *Message) {
	now := time.Now().UTC().Unix()
	msg = &Message{
		Payload: Payload{
			Issuer:    sender.ID,
			Audience:  receiver.ID,
			Subject:   subject,
			IssuedAt:  now,
			NotBefore: now,
			ExpiredAt: now + DefaultTTL,
		},
		Footer: Footer{
			PeerID: sender.ID,
		},
	}
	return msg
}

// Unpack returns the decoded token into a [Message].
func (msg *Message) Unpack(header, token string, implicit []byte) (err error) {
	logp := `Unpack`

	// Step 3: verify the header and unpack the footer if it exists.
	token, found := strings.CutPrefix(token, header)
	if !found {
		return fmt.Errorf(`%s: invalid header, want %s`, logp, header)
	}
	token, footerb64, found := strings.Cut(token, `.`)
	if found {
		msg.RawFooter, err = base64.RawURLEncoding.DecodeString(footerb64)
		if err != nil {
			return fmt.Errorf(`%s: invalid footer: %w`, logp, err)
		}
	}

	// Step 4: Decodes the payload.
	paysig, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	lenpaysig := len(paysig)
	if lenpaysig <= ed25519.SignatureSize {
		return fmt.Errorf(`%s: invalid payload size %d`, logp, lenpaysig)
	}
	msg.RawPayload = paysig[:lenpaysig-64]
	msg.Sig = paysig[lenpaysig-64:]

	// Step 5: Generate PAE.
	msg.PAE, err = PreAuthEncode([]byte(header), msg.RawPayload,
		msg.RawFooter, implicit)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}
