// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
)

var (
	_resPool = sync.Pool{
		New: func() interface{} {
			return new(Response)
		},
	}
)

// Response contains the data that server send to client as a reply from
// Request or as broadcast from client subscription.
//
// If response type is a reply from Request, the ID from Request will be
// copied to the Response, and `code` and `message` field values are equal
// with HTTP response code and message.
//
// Example of response format for replying request,
//
//	{
//		id: 1512459721269,
//		code:  200,
//		message: "",
//		body: ""
//	}
//
// If response type is broadcast the ID and code MUST be 0, and the `message`
// field will contain the name of subscription. For example, when recipient of
// message read the message, server will publish a notification response as,
//
//	{
//		id: 0,
//		code:  0,
//		message: "message.read",
//		body: "{ \"id\": ... }"
//	}
type Response struct {
	Message string `json:"message"`
	Body    string `json:"body"`
	ID      uint64 `json:"id"`
	Code    int32  `json:"code"`
}

// NewBroadcast create a new message for broadcast by server encoded as JSON
// and wrapped in TEXT frame.
func NewBroadcast(message, body string) (packet []byte, err error) {
	var (
		logp = `NewBroadcast`
		res  = _resPool.Get().(*Response)
	)

	res.reset()

	res.Message = message
	res.Body = body

	packet, err = json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	_resPool.Put(res)

	packet = NewFrameText(false, packet)

	return packet, nil
}

// reset all field's value to zero or empty.
func (res *Response) reset() {
	res.ID = 0
	res.Code = 0
	res.Message = ""
	res.Body = ""
}
