// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package smtp

import "strings"

// LocalHandler is an handler using local environment.
type LocalHandler struct {
	env *Environment
}

// NewLocalHandler create an handler using local environment.
func NewLocalHandler(env *Environment) (local *LocalHandler) {
	local = &LocalHandler{
		env: env,
	}
	return local
}

// ServeAuth handle SMTP AUTH parameter username and password.
func (lh *LocalHandler) ServeAuth(username, password string) (
	res *Response, err error,
) {
	res = &Response{
		Code:    StatusAuthenticated,
		Message: "2.7.0 Authentication successful",
	}

	username = strings.ToLower(username)
	localDomain := strings.Split(username, "@")
	if len(localDomain) != 2 {
		return nil, ErrInvalidCredential
	}

	if localDomain[1] == lh.env.PrimaryDomain.Name {
		acc, ok := lh.env.PrimaryDomain.Accounts[localDomain[0]]
		if !ok {
			return nil, ErrInvalidCredential
		}

		// System accounts have empty length password.
		if len(acc.HashPass) == 0 {
			return nil, ErrInvalidCredential
		}

		err = acc.Authenticate(password)
		if err != nil {
			return nil, ErrInvalidCredential
		}

		return res, nil
	}

	for _, vdom := range lh.env.VirtualDomains {
		if localDomain[1] != vdom.Name {
			continue
		}

		acc, ok := vdom.Accounts[localDomain[0]]
		if !ok {
			return nil, ErrInvalidCredential
		}
		// System accounts have empty length password.
		if len(acc.HashPass) == 0 {
			return nil, ErrInvalidCredential
		}

		err = acc.Authenticate(password)
		if err != nil {
			return nil, ErrInvalidCredential
		}

		return res, nil
	}

	return nil, ErrInvalidCredential
}

// ServeBounce handle email transaction with unknown or invalid recipent.
func (lh *LocalHandler) ServeBounce(_ *MailTx) (res *Response, err error) {
	// TODO: send delivery status notification to sender address.
	return nil, nil
}

// ServeExpand handle SMTP EXPN command.
//
// TODO: The group feature currently is not supported.
func (lh *LocalHandler) ServeExpand(_ string) (res *Response, err error) {
	res = &Response{
		Code:    StatusCmdNotImplemented,
		Message: "Command not implemented",
	}
	return res, nil
}

// ServeMailTx handle processing the final delivery of incoming mail.
// TODO: implement it.
func (lh *LocalHandler) ServeMailTx(_ *MailTx) (res *Response, err error) {
	return nil, nil
}

// ServeVerify handle SMTP VRFY command.  The username must be in the format
// of mailbox, "local@domain".
func (lh *LocalHandler) ServeVerify(username string) (res *Response, err error) {
	username = strings.ToLower(username)
	localDomain := strings.Split(username, "@")
	if len(localDomain) != 2 {
		return nil, errCmdSyntaxError
	}

	res = &Response{
		Code: StatusOK,
	}

	if localDomain[1] == lh.env.PrimaryDomain.Name {
		acc, ok := lh.env.PrimaryDomain.Accounts[localDomain[0]]
		if ok {
			res.Message = acc.String()
			return res, nil
		}
	}
	for _, vdom := range lh.env.VirtualDomains {
		if localDomain[1] != vdom.Name {
			continue
		}
		acc, ok := vdom.Accounts[localDomain[0]]
		if ok {
			res.Message = acc.String()
			return res, nil
		}
	}

	res.Code = StatusMailboxNotFound
	res.Message = "mailbox not found"

	return res, nil
}
