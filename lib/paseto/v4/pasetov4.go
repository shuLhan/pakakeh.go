// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

// Package pasetov4 provides a simple, ready to use, opinionated
// implementation of Platform-Agnostic SEcurity TOkens (PASETO) version 4 as
// defined in [paseto-v4].
//
// There are two protocols in paseto v4: local and public.
//
// # Local mode
//
// Local mode use symmetric key (a secret), it works by encrypting and
// decrypting message, see [LocalMode.Encrypt] and [LocalMode.Decrypt].
//
// # Public mode
//
// Public mode use asymmetric key (ed25519 private and public key).
// Public mode works by signing and verifying message,
// see [PublicMode.Sign] and [PublicMode.Verify], similar to the standard
// specification.
//
// # PublicMode as JWT
//
// The opinioted implementation of Paseto v4 public protocol sends and
// receives [paseto.Message] using [PublicMode.Pack] and [PublicMode.Unpack]
// with one or more [paseto.Peer].
//
// To illustrate how they works, let say we have a sender and a receiver.
// Sender and receiver has their own asymmetric key and unique ID.
// Sender send a message to receiver.
//
// # Sender side
//
// On sender side, it works this way,
//
// 1) Initialize public mode with seed,
//
//	senderpm := pasetov4.NewPublicMode(`senderID`, senderSeed)
//	receiver := Peer{
//		ID: `receiverID`,
//		PublicKey: receiverPublicKey,
//	}
//
// 2) Create message using sender and receiver with optional subject,
//
//	msg := paseto.NewMessage(senderpm.Peer, receiver, `hello`)
//
// This will fill the JWT claims inside the [paseto.Message.Payload], excepts
// for [paseto.Payload.Data] and [paseto.Footer.Data].
//
//	Message{
//		Payload: {
//			Issuer: "senderID",
//			Audience: "receiverID",
//			Subject: "hello",
//			IssuedAt: <current-timestamp>,
//			NotBefore: <current-timestamp>,
//			ExpiredAt: <current-timestamp> + [paseto.DefaultTTL],
//		},
//		Footer: {
//			PeerID: "senderID"
//		},
//	}
//
// 3) Set the payload and optional footer data,
//
//	msg.Payload.Data = anydata
//	msg.Footer.Data = optdata
//
// 4) Pack the message to generate new public token
//
//	token, _ := pmode.Pack(msg, nil)
//
// 5) Now you can send the token to the receiver.
//
// # Receiver side
//
// Steps to use public mode on receiver side,
//
// 1) Initialize public mode with seed,
//
//	receiverpm := pasetov4.NewPublicMode(`senderID`, senderSeed)
//	sender := Peer{
//		ID: `senderID`,
//		PublicKey: senderPublicKey,
//	}
//
// 2) Add sender as known peer,
//
//	receiverpm.AddPeer(sender)
//
// 3) Unpack the token received from sender,
//
//	msg := Message{}
//	receiverpm.Unpack(token, nil, &msg)
//
// The Unpack method verify the token using the sender public key.
// If the token is verified it will decode the payload and footer into the msg
// variable.
//
// 4) Validate the payload to check if its valid (belong to receiver and not
// expired),
//
//	msg.Payload.Validate(receiver.Peer.ID, sender)
//
// [paseto-v4]: https://github.com/paseto-standard/paseto-spec/blob/master/docs/01-Protocol-Versions/Version4.md
package pasetov4
