// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

//
// Response contains DNS reply message for client.
//
type Response struct {
	ReceivedAt int
	Message    *Message
}
