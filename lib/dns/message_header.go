// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	headerIsQuery    byte = 0x00
	headerIsResponse byte = 0x80 // 1000.0000
	headerMaskOpCode byte = 0x78 // 0111.1000
	headerIsAA       byte = 0x04 // 0000.0100
	headerIsTC       byte = 0x02 // 0000.0010
	headerIsRD       byte = 0x01 // 0000.0001
	headerIsRA       byte = 0x80 //          1000.0000
	headerMaskRCode  byte = 0x0F //          0000.1111
)

// MessageHeader the header includes fields that specify which of the
// remaining sections are present, and also specify whether the message is a
// query or a response, a standard query or some other opcode, etc. [1]
//
// The header section is always present.
//
// [1] RFC 1035 P-25 - 4.1. Format
type MessageHeader struct {
	//
	// A 16 bit identifier assigned by the program that generates
	// any kind of query.  This identifier is copied the corresponding
	// reply and can be used by the requester to match up replies to
	// outstanding queries.
	//
	ID uint16

	//
	// A one bit field that specifies whether this message is a query (0),
	// or a response (1).
	//
	IsQuery bool

	//
	// A four bit field that specifies kind of query in this message.
	// This value is set by the originator of a query and copied into the
	// response.
	//
	Op OpCode

	//
	// Authoritative Answer - this bit is valid in responses, and
	// specifies that the responding name server is an authority for the
	// domain name in question section.  Note that the contents of the
	// answer section may have multiple owner names because of aliases.
	// The AA bit corresponds to the name which matches the query name, or
	// the first owner name in the answer section.
	//
	IsAA bool

	//
	// TrunCation - specifies that this message was truncated due to
	// length greater than that permitted on the transmission channel.
	//
	IsTC bool

	//
	// Recursion Desired - this bit may be set in a query and is copied
	// into the response.  If RD is set, it directs the name server to
	// pursue the query recursively.  Recursive query support is optional.
	//
	IsRD bool

	//
	// Recursion Available - this bit is set or cleared in a response, and
	// denotes whether recursive query support is available in the name
	// server.
	//
	IsRA bool

	//
	// Response code - this 4 bit field is set as part of responses.
	//
	RCode ResponseCode

	// The number of entries in the question section.
	QDCount uint16

	// The number of resource records in the answer section.
	ANCount uint16

	// The number of name server resource records in the authority records
	// section.
	NSCount uint16

	// The number of resource records in the additional records section.
	ARCount uint16
}

// Reset the header to default (query) values, which mean the IsQuery is true,
// the Op code is 0, with recursion enabled, and query count set tot 1.
func (hdr *MessageHeader) Reset() {
	hdr.ID = 0
	hdr.IsQuery = true
	hdr.Op = OpCodeQuery
	hdr.IsAA = false
	hdr.IsTC = false
	hdr.IsRD = true
	hdr.IsRA = false
	hdr.RCode = RCodeOK
	hdr.QDCount = 1
	hdr.ANCount = 0
	hdr.NSCount = 0
	hdr.ARCount = 0
}

// pack the section header into slice of bytes.
func (hdr *MessageHeader) pack() []byte {
	var (
		b0, b1 byte
		packet [12]byte
	)

	packet[0] = byte(hdr.ID >> 8)
	packet[1] = byte(hdr.ID)

	if hdr.IsQuery {
		b0 = headerIsQuery
	} else {
		b0 = headerIsResponse
	}

	b0 |= (headerMaskOpCode & byte(hdr.Op<<3))

	if hdr.IsRD {
		b0 |= headerIsRD
	}

	if !hdr.IsQuery {
		if hdr.IsAA {
			b0 |= headerIsAA
		}
		if hdr.IsTC {
			b0 |= headerIsTC
		}
		if hdr.IsRA {
			b1 |= headerIsRA
		}
		b1 |= (headerMaskRCode & byte(hdr.RCode))
	}

	packet[2] = b0
	packet[3] = b1

	binary.BigEndian.PutUint16(packet[4:], hdr.QDCount)
	binary.BigEndian.PutUint16(packet[6:], hdr.ANCount)
	binary.BigEndian.PutUint16(packet[8:], hdr.NSCount)
	binary.BigEndian.PutUint16(packet[10:], hdr.ARCount)

	return packet[:]
}

// unpack the DNS header section.
func (hdr *MessageHeader) unpack(packet []byte) (err error) {
	if len(packet) < sectionHeaderSize {
		return errors.New(`header too small`)
	}
	hdr.Op = OpCode((packet[2] & headerMaskOpCode) >> 3)
	if hdr.Op < 0 || hdr.Op > OpCodeStatus {
		return fmt.Errorf(`unknown op code=%d`, hdr.Op)
	}
	hdr.RCode = ResponseCode(headerMaskRCode & packet[3])
	if hdr.RCode < 0 || hdr.RCode > RCodeRefused {
		return fmt.Errorf(`unknown response code=%d`, hdr.RCode)
	}

	hdr.ID = binary.BigEndian.Uint16(packet)

	hdr.IsQuery = packet[2]&headerIsResponse != headerIsResponse
	hdr.IsAA = packet[2]&headerIsAA == headerIsAA
	hdr.IsTC = packet[2]&headerIsTC == headerIsTC
	hdr.IsRD = packet[2]&headerIsRD == headerIsRD
	hdr.IsRA = packet[3]&headerIsRA == headerIsRA

	hdr.QDCount = binary.BigEndian.Uint16(packet[4:])
	hdr.ANCount = binary.BigEndian.Uint16(packet[6:])
	hdr.NSCount = binary.BigEndian.Uint16(packet[8:])
	hdr.ARCount = binary.BigEndian.Uint16(packet[10:])

	return nil
}
