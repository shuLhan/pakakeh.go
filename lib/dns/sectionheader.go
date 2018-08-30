// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	libbytes "github.com/shuLhan/share/lib/bytes"
)

const (
	headerIsQuery    byte = 0x00
	headerIsResponse byte = 0x80
	headerIsAA       byte = 0x04
	headerIsTC       byte = 0x02
	headerIsRD       byte = 0x01
	headerIsRA       byte = 0x80
)

//
// SectionHeader The header section is always present.  The header includes
// fields that specify which of the remaining sections are present, and also
// specify whether the message is a query or a response, a standard query or
// some other opcode, etc. [1]
//
// [1] RFC 1035 P-25 - 4.1. Format
//
type SectionHeader struct {
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

	// An unsigned 16 bit integer specifying the number of entries in the
	// question section.
	QDCount uint16

	// An unsigned 16 bit integer specifying the number of resource
	// records in the answer section.
	ANCount uint16

	// An unsigned 16 bit integer specifying the number of name server
	// resource records in the authority records section.
	NSCount uint16

	// An unsigned 16 bit integer specifying the number of resource
	// records in the additional records section.
	ARCount uint16
}

//
// Reset the header to default (query) values.
//
func (hdr *SectionHeader) Reset() {
	hdr.ID = 0
	hdr.IsQuery = true
	hdr.Op = OpCodeQuery
	hdr.IsAA = false
	hdr.IsTC = false
	hdr.IsRD = false
	hdr.IsRA = false
	hdr.RCode = RCodeOK
	hdr.QDCount = 0
	hdr.ANCount = 0
	hdr.NSCount = 0
	hdr.ARCount = 0
}

//
// MarshalBinary pack the section header into slice of bytes.
//
func (hdr *SectionHeader) MarshalBinary() ([]byte, error) {
	var b0, b1 byte

	packet := make([]byte, 4)

	packet[0] = byte(hdr.ID >> 8)
	packet[1] = byte(hdr.ID)

	if hdr.IsQuery {
		b0 = headerIsQuery
	} else {
		b0 = headerIsResponse
	}

	b0 = b0 | (0x78 & byte(hdr.Op<<2))

	if hdr.IsRD {
		b0 = b0 | headerIsRD
	}

	if !hdr.IsQuery {
		if hdr.IsAA {
			b0 = b0 | headerIsAA
		}
		if hdr.IsTC {
			b0 = b0 | headerIsTC
		}
		if hdr.IsRA {
			b1 = b1 | headerIsRA
		}
		b1 = b1 | (0x0F & byte(hdr.RCode))
	}

	packet[2] = b0
	packet[3] = b1

	libbytes.AppendUint16(&packet, hdr.QDCount)
	libbytes.AppendUint16(&packet, hdr.ANCount)
	libbytes.AppendUint16(&packet, hdr.NSCount)
	libbytes.AppendUint16(&packet, hdr.ARCount)

	return packet, nil
}

//
// UnmarshalBinary unpack the DNS header section.
//
func (hdr *SectionHeader) UnmarshalBinary(packet []byte) error {
	hdr.ID = libbytes.ReadUint16(packet, 0)

	if packet[2]&headerIsResponse == headerIsResponse {
		hdr.IsQuery = false
	}
	hdr.Op = OpCode((packet[2] & 0x78) >> 2)

	if packet[2]&headerIsAA == headerIsAA {
		hdr.IsAA = true
	}
	if packet[2]&headerIsTC == headerIsTC {
		hdr.IsTC = true
	}
	if packet[2]&headerIsRD == headerIsRD {
		hdr.IsRD = true
	}
	if packet[3]&headerIsRA == headerIsRA {
		hdr.IsRA = true
	}

	hdr.RCode = ResponseCode(0x0F & packet[3])

	hdr.QDCount = libbytes.ReadUint16(packet, 4)
	hdr.ANCount = libbytes.ReadUint16(packet, 6)
	hdr.NSCount = libbytes.ReadUint16(packet, 8)
	hdr.ARCount = libbytes.ReadUint16(packet, 10)

	return nil
}
