// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

type HandlerPosix struct{}

func (hp *HandlerPosix) ServeAuth(username, password string) (
	res *Response, err error,
) {
	return nil, nil
}

func (hp *HandlerPosix) ServeBounce(mail *MailTx) (res *Response, err error) {
	return nil, nil
}

func (hp *HandlerPosix) ServeExpand(mailingList string) (res *Response, err error) {
	return nil, nil
}

//
// ServeMailTx handle processing the final delivery of incoming mail.
//
func (hp *HandlerPosix) ServeMailTx(mail *MailTx) (res *Response, err error) {
	return nil, nil
}

func (hp *HandlerPosix) ServeVerify(username string) (res *Response, err error) {
	return nil, nil
}
