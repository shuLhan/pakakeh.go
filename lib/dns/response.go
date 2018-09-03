// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"time"
)

//
// Response contains DNS reply message for client.
//
type Response struct {
	ReceivedAt int64
	Message    *Message
}

//
// Reset response to empty state.
//
func (res *Response) Reset() {
	res.ReceivedAt = 0
	res.Message.Reset()
}

//
// IsExpired will return true if response message is expired, otherwise it
// will return false.
//
func (res *Response) IsExpired() bool {
	// Local responses from hosts file will never be expired.
	if res.ReceivedAt == 0 {
		return false
	}

	elapSeconds := uint32(time.Now().Unix() - res.ReceivedAt)

	return res.Message.IsExpired(elapSeconds)
}

//
// Unpack message and set received time value to current time.
//
func (res *Response) Unpack() (err error) {
	err = res.Message.Unpack()
	if err != nil {
		return
	}

	res.ReceivedAt = time.Now().Unix()

	return
}
