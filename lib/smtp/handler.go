// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Handler define an interface to handle bouncing and incoming mail message,
// and handling EXPN and VRFY commands.
//
type Handler interface {
	ServeAuth(username, password string) (*Response, error)
	ServeBounce(mail *MailTx) (*Response, error)
	ServeExpand(mailingList string) (*Response, error)
	ServeMailTx(mail *MailTx) (*Response, error)
	ServeVerify(username string) (*Response, error)
}
