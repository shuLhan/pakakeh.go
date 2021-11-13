// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"container/list"
	"time"
)

//
// Answer maintain the record of DNS response for cache.
//
type Answer struct {
	// ReceivedAt contains time when message is received.  If answer is
	// from local cache (host or zone file), its value is 0.
	ReceivedAt int64

	// AccessedAt contains time when message last accessed.  This field
	// is used to prune old answer from caches.
	AccessedAt int64

	// QName contains DNS question name, a copy of msg.Question.Name.
	QName string
	// qtype contains DNS question type, a copy of msg.Question.Type.
	QType uint16
	// qclass contains DNS question class, a copy of msg.Question.Class.
	QClass uint16

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
func newAnswer(msg *Message, isLocal bool) (an *Answer) {
	an = &Answer{
		QName:  msg.Question.Name,
		QType:  msg.Question.Type,
		QClass: msg.Question.Class,
		msg:    msg,
	}
	if isLocal {
		return
	}
	at := time.Now().Unix()
	an.ReceivedAt = at
	an.AccessedAt = at
	return
}

//
// clear the answer fields.
//
func (an *Answer) clear() {
	an.msg = nil
	an.el = nil
}

//
// get the raw packet in the message.
// Before the raw packet is returned, the answer accessed time will be updated
// to current time and each resource record's TTL in message is subtracted
// based on received time.
//
func (an *Answer) get() (packet []byte) {
	an.updateTTL()

	packet = make([]byte, len(an.msg.packet))
	copy(packet, an.msg.packet)
	return
}

//
// update the answer with new message.
//
func (an *Answer) update(nu *Answer) {
	if nu == nil || nu.msg == nil {
		return
	}

	if an.ReceivedAt > 0 {
		an.ReceivedAt = nu.ReceivedAt
		an.AccessedAt = nu.AccessedAt
	}

	an.msg = nu.msg
	nu.msg = nil
}

//
// updateTTL decrease the answer TTLs based on time when message received.
//
func (an *Answer) updateTTL() {
	if an.ReceivedAt == 0 {
		return
	}

	an.AccessedAt = time.Now().Unix()
	ttl := uint32(an.AccessedAt - an.ReceivedAt)
	an.msg.SubTTL(ttl)
}
