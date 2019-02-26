// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Handler define an interface to handle bouncing and incoming mail message,
// and handling EXPN and VRFY commands.
//
type Handler interface {
	// ServeAuth handle SMTP AUTH parameter username and password.
	ServeAuth(username, password string) (*Response, error)

	// ServeBounce handle email transaction that with unknown or invalid
	// recipent.
	ServeBounce(mail *MailTx) (*Response, error)

	// ServeExpand handle SMTP EXPN command.
	ServeExpand(mailingList string) (*Response, error)

	// ServeMailTx handle termination on email transaction.
	ServeMailTx(mail *MailTx) (*Response, error)

	// ServeVerify handle SMTP VRFY command.
	ServeVerify(username string) (*Response, error)
}
