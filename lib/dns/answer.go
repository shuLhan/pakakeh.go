// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"container/list"
	"time"
)

//
// answer maintain the record of DNS response for cache.
//
type answer struct {
	// receivedAt contains time when message is received.  If answer is
	// from local cache (host or zone file), its value is 0.
	receivedAt int64

	// accessedAt contains time when message last accessed.  This field
	// is used to prune old answer from caches.
	accessedAt int64

	// qname contains DNS question name, a copy of msg.Question.Name.
	qname string
	// qtype contains DNS question type, a copy of msg.Question.Type.
	qtype uint16
	// qclass contains DNS question class, a copy of msg.Question.Class.
	qclass uint16

	// msg contains the unpacked DNS message.
	msg *Message

	// el contains pointer to the cache in LRU.
	el *list.Element
}

//
// newAnswer create new answer from Message.
// If is not local (isLocal=false), the received and accessed time will be set
// to current timestamp.
//
func newAnswer(msg *Message, isLocal bool) (an *answer) {
	if msg == nil || msg.Question == nil || len(msg.Answer) == 0 {
		return nil
	}

	an = &answer{
		qname:  string(msg.Question.Name),
		qtype:  msg.Question.Type,
		qclass: msg.Question.Class,
		msg:    msg,
	}
	if isLocal {
		return
	}
	at := time.Now().Unix()
	an.receivedAt = at
	an.accessedAt = at
	return
}

//
// clear the answer fields.
//
func (an *answer) clear() {
	an.msg = nil
	an.el = nil
}

//
// get the raw packet in the message.
// Before the raw packet is returned, the answer accessed time will be updated
// to current time and each resource record's TTL in message is subtracted
// based on received time.
//
func (an *answer) get() (packet []byte) {
	if an.receivedAt > 0 {
		an.accessedAt = time.Now().Unix()
		ttl := uint32(an.accessedAt - an.receivedAt)
		an.msg.SubTTL(ttl)
	}

	packet = make([]byte, len(an.msg.Packet))
	copy(packet, an.msg.Packet)
	return
}

//
// update the answer with new message.
//
func (an *answer) update(nu *answer) {
	if nu == nil || nu.msg == nil {
		return
	}

	if an.receivedAt > 0 {
		an.receivedAt = nu.receivedAt
		an.accessedAt = nu.accessedAt
	}

	an.msg = nu.msg
	nu.msg = nil
}
