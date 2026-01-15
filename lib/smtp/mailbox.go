// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package smtp

// Mailbox represent a mailbox format.
type Mailbox struct {
	Name   string // Name of user in system.
	Local  string // Local part.
	Domain string // Domain part.
}
