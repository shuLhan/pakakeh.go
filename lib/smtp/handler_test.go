// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

type testHandler struct{}

func (th *testHandler) ServeAuth(username, password string) (
	res *Response, err error,
) {
	if username == testUsername && password == testPassword {
		res = &Response{
			Code:    StatusAuthenticated,
			Message: "2.7.0 Authentication successful",
		}
		return res, nil
	}
	return nil, ErrInvalidCredential
}

func (th *testHandler) ServeBounce(mail *MailTx) (res *Response, err error) {
	res = &Response{
		Code: StatusOK,
	}
	return res, nil
}

func (th *testHandler) ServeExpand(mailingList string) (res *Response, err error) {
	res = &Response{}
	if mailingList == "list-exist" {
		res.Code = StatusOK
		res.Message = "List Exist"
		res.Body = []string{
			"Member A <member-a@mail.local>",
		}
	}
	return res, nil
}

//
// ServeMailTx handle processing the final delivery of incoming mail.
//
func (th *testHandler) ServeMailTx(mail *MailTx) (res *Response, err error) {
	res = &Response{
		Code: StatusOK,
	}
	return res, nil
}

func (th *testHandler) ServeVerify(username string) (res *Response, err error) {
	switch username {
	case "exist":
		res = &Response{
			Code:    StatusOK,
			Message: "Exist <exist@mail.local>",
		}
	case "notexist":
		res = &Response{
			Code:    StatusMailboxNotFound,
			Message: "No such user here",
		}
	case "ambigous":
		res = &Response{
			Code:    StatusMailboxIncorrect,
			Message: "User ambigous",
			Body: []string{
				"Ambigous A <a@mail.local>",
				"Ambigous B <b@mail.local>",
				"Ambigous C <c@mail.local>",
			},
		}
	}
	return res, nil
}
