// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

// Message defines the payload be signed and verified by sender/receiver.
type Message struct {
	Payload Payload
	Footer  Footer
}
