// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dns

import (
	"container/list"
	"fmt"
	"strings"
	"time"
)

// Answer maintain the record of DNS response for cache.
type Answer struct {
	// el contains pointer to the cache in LRU.
	el *list.Element

	// Message contains the unpacked DNS message.
	Message *Message

	// QName contains DNS question name, a copy of msg.Question.Name.
	QName string

	// ReceivedAt contains time when message is received.
	// A zero value indicated local answer (loaded from hosts or zone
	// files).
	ReceivedAt int64

	// AccessedAt contains time when message last accessed.
	// This field is used to prune old answer from caches.
	AccessedAt int64

	// RType contains record type, a copy of msg.Question.Type.
	RType RecordType

	// RClass contains record class, a copy of msg.Question.Class.
	RClass RecordClass

	// TTL contains the first TTL on RR Answer.
	TTL uint32
}

// newAnswer create new answer from Message.
// If is not local (isLocal=false), the received and accessed time will be set
// to current timestamp.
func newAnswer(msg *Message, isLocal bool) (an *Answer) {
	an = &Answer{
		// Trim the dot at the end for Message that is come from zone.
		QName:   strings.TrimSuffix(msg.Question.Name, `.`),
		RType:   msg.Question.Type,
		RClass:  msg.Question.Class,
		Message: msg,
	}
	if isLocal {
		return
	}
	var at = time.Now().Unix()
	an.ReceivedAt = at
	an.AccessedAt = at
	if len(msg.Answer) != 0 {
		an.TTL = msg.Answer[0].TTL
	}
	return
}

func (an *Answer) String() string {
	var id uint16
	if an.Message != nil {
		id = an.Message.Header.ID
	}
	return fmt.Sprintf(`{%d %s %s %d}`, id, an.QName,
		RecordTypeNames[an.RType], an.TTL)
}

// clear the answer fields.
func (an *Answer) clear() {
	an.Message = nil
	an.el = nil
}

// get the raw packet in the message.
// Before the raw packet is returned, the answer accessed time will be updated
// to current time and each resource record's TTL in message is subtracted
// based on received time.
func (an *Answer) get() (packet []byte) {
	an.updateTTL()

	packet = make([]byte, len(an.Message.packet))
	copy(packet, an.Message.packet)
	return
}

// update the answer with new message.
func (an *Answer) update(nu *Answer) {
	if nu == nil || nu.Message == nil {
		return
	}

	if an.ReceivedAt > 0 {
		an.ReceivedAt = nu.ReceivedAt
		an.AccessedAt = nu.AccessedAt
	}

	an.Message = nu.Message
	an.TTL = nu.TTL
}

// updateTTL decrease the answer TTLs based on time when message received.
func (an *Answer) updateTTL() {
	if an.ReceivedAt == 0 {
		return
	}

	an.AccessedAt = time.Now().Unix()
	an.TTL = uint32(an.AccessedAt - an.ReceivedAt)
	an.Message.SubTTL(an.TTL)
}
