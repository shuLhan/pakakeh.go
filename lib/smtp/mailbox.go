// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

// Mailbox represent a mailbox format.
type Mailbox struct {
	Name   string // Name of user in system.
	Local  string // Local part.
	Domain string // Domain part.
}
