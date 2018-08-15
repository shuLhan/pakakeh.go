// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
)

//
// MessageQueue define list of message queried by clients to name servers.
//
type MessageQueue []*Message

//
// DefMsgQueue default message queue for all clients.
//
var DefMsgQueue MessageQueue

//
// Push a message to queue.
//
func (mq MessageQueue) Push(msg *Message) {
	mq = append(mq, msg)
}

//
// Pop a message in queue matched by ID, type, class, and name.  If none are
// matched, it will return nil.
//
func (mq MessageQueue) Pop(msg *Message) *Message {
	for x := 0; x < len(mq); x++ {
		if mq[x].Header.ID != msg.Header.ID {
			continue
		}
		if mq[x].Question.Type != msg.Question.Type {
			continue
		}
		if mq[x].Question.Class != msg.Question.Class {
			continue
		}
		if !bytes.Equal(mq[x].Question.Name, msg.Question.Name) {
			continue
		}

		msgQuery := mq[x]

		mq = append(mq[:x], mq[x+1:]...)

		return msgQuery
	}

	return nil
}
