// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>

package paseto

// Footer defines the peer ID and optional data for the message.
// Footer is not included during signing payload, it can be tampered, so it
// MUST NOT include sensitive information.
type Footer struct {
	Data   any    `json:"data,omitempty"`
	PeerID string `json:"peer_id,omitempty"`
}
