// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"log"
)

//
// Message represent a DNS message.
//
// All communications inside of the domain protocol are carried in a single
// format called a message.  The top level format of message is divided
// into 5 sections (some of which are empty in certain cases) shown below:
//
//     +---------------------+
//     |        Header       |
//     +---------------------+
//     |       Question      | the question for the name server
//     +---------------------+
//     |        Answer       | RRs answering the question
//     +---------------------+
//     |      Authority      | RRs pointing toward an authority
//     +---------------------+
//     |      Additional     | RRs holding additional information
//     +---------------------+
//
// The names of the sections after the header are derived from their use in
// standard queries.  The question section contains fields that describe a
// question to a name server.  These fields are a query type (QTYPE), a
// query class (QCLASS), and a query domain name (QNAME).  The last three
// sections have the same format: a possibly empty list of concatenated
// resource records (RRs).  The answer section contains RRs that answer the
// question; the authority section contains RRs that point toward an
// authoritative name server; the additional records section contains RRs
// which relate to the query, but are not strictly answers for the
// question. [1]
//
// [1] RFC 1035 - 4.1. Format
//
type Message struct {
	Header     *SectionHeader
	Question   *SectionQuestion
	Answer     []*ResourceRecord
	Authority  []*ResourceRecord
	Additional []*ResourceRecord

	// Slice that hold the result of packing the message or original
	// message from unpacking.
	Packet []byte
}

//
// Reset the message fields.
//
func (msg *Message) Reset() {
	msg.Header.Reset()
	msg.Question.Reset()

	for x := 0; x < len(msg.Answer); x++ {
		rrPool.Put(msg.Answer[x])
	}
	for x := 0; x < len(msg.Authority); x++ {
		rrPool.Put(msg.Authority[x])
	}
	for x := 0; x < len(msg.Additional); x++ {
		rrPool.Put(msg.Additional[x])
	}

	msg.Answer = msg.Answer[:0]
	msg.Authority = msg.Authority[:0]
	msg.Additional = msg.Additional[:0]
	msg.Packet = msg.Packet[:0]
}

func (msg *Message) writeUint16(x uint16) {
	msg.Packet = append(msg.Packet, byte(x>>8))
	msg.Packet = append(msg.Packet, byte(0x00FF&x))
}

func (msg *Message) writeLabel(label []byte) {
	var count byte

	idx := len(msg.Packet)
	msg.Packet = append(msg.Packet, count)

	for x := 0; x < len(label); x++ {
		if label[x] == '.' {
			// Skip label that prefixed with '.', e.g.
			// '...test.com'
			if count == 0 {
				continue
			}

			msg.Packet[idx] = count
			count = 0
			idx = len(msg.Packet)
			msg.Packet = append(msg.Packet, count)
			continue
		}

		msg.Packet = append(msg.Packet, label[x])
		count++
	}
	if count > 0 {
		msg.Packet[idx] = count
		count = 0
	}
	msg.Packet = append(msg.Packet, count)
}

//
// MarshalBinary convert message into datagram packet.  The result of packing a message
// will be saved in Packet field.
//
func (msg *Message) MarshalBinary() ([]byte, error) {
	var b0, b1 byte

	msg.Packet = msg.Packet[:0]

	msg.Packet = append(msg.Packet, byte(msg.Header.ID>>8))
	msg.Packet = append(msg.Packet, byte(msg.Header.ID))

	if msg.Header.IsQuery {
		b0 = HeaderIsQuery
	} else {
		b0 = HeaderIsResponse
	}

	b0 = b0 | (0x78 & byte(msg.Header.Op<<2))

	if msg.Header.IsQuery {
		if msg.Header.IsRD {
			b0 = b0 | HeaderIsRD
		}
	} else {
		if msg.Header.IsAA {
			b0 = b0 | HeaderIsAA
		}
		if msg.Header.IsTC {
			b0 = b0 | HeaderIsTC
		}
		if msg.Header.IsRA {
			b1 = b1 | HeaderIsRA
		}
		b1 = b1 | (0x0F & byte(msg.Header.RCode))
	}

	msg.Packet = append(msg.Packet, b0)
	msg.Packet = append(msg.Packet, b1)

	msg.writeUint16(msg.Header.QDCount)

	if msg.Header.IsQuery {
		msg.writeUint16(0)
		msg.writeUint16(0)
		msg.writeUint16(0)
	} else {
		msg.writeUint16(msg.Header.ANCount)
		msg.writeUint16(msg.Header.NSCount)
		msg.writeUint16(msg.Header.ARCount)
	}

	msg.writeLabel(msg.Question.Name)
	msg.writeUint16(uint16(msg.Question.Type))
	msg.writeUint16(uint16(msg.Question.Class))

	if msg.Header.IsQuery {
		return msg.Packet, nil
	}

	return msg.Packet, nil
}

//
// UnmarshalBinary unpack the packet to fill the message fields.
//
func (msg *Message) UnmarshalBinary(packet []byte) error {
	_ = msg.Header.UnmarshalBinary(packet)

	if debugLevel >= 1 {
		log.Printf("msg.Header: %+v\n", msg.Header)
	}

	if len(packet) <= sectionHeaderSize {
		return nil
	}

	err := msg.Question.UnmarshalBinary(packet[12:])
	if err != nil {
		return err
	}

	if debugLevel >= 1 {
		log.Printf("msg.Question: %s\n", msg.Question)
	}

	startIdx := sectionHeaderSize + msg.Question.Size()

	var x uint16
	for ; x < msg.Header.ANCount; x++ {
		rr := rrPool.Get().(*ResourceRecord)
		rr.Reset()
		startIdx, err = rr.Unpack(packet, startIdx)
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
	}

	if debugLevel >= 1 {
		log.Printf("msg.Answer: %+v\n", msg.Answer)
	}

	for x = 0; x < msg.Header.NSCount; x++ {
		rr := rrPool.Get().(*ResourceRecord)
		rr.Reset()
		startIdx, err = rr.Unpack(packet, startIdx)
		if err != nil {
			return err
		}
		msg.Authority = append(msg.Authority, rr)
	}

	if debugLevel >= 1 {
		log.Printf("msg.Authority: %+v\n", msg.Authority)
	}

	for x = 0; x < msg.Header.ARCount; x++ {
		rr := rrPool.Get().(*ResourceRecord)
		rr.Reset()
		startIdx, err = rr.Unpack(packet, startIdx)
		if err != nil {
			return err
		}
		msg.Additional = append(msg.Additional, rr)
	}

	if debugLevel >= 1 {
		log.Printf("msg.Additional: %+v\n", msg.Additional)
	}

	return nil
}
