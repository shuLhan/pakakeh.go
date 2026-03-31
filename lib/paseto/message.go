// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"time"
)

// Message defines the payload be signed and verified by sender/receiver.
type Message struct {
	// The following fields are filled after Unpack and MUST not be used
	// directly.
	RawPayload []byte `json:"-"`
	RawFooter  []byte `json:"-"`
	Sig        []byte `json:"-"`
	PAE        []byte `json:"-"`

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
