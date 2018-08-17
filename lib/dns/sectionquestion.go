// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// SectionQuestion The question section is used to carry the "question" in
// most queries, i.e., the parameters that define what is being asked.  The
// section contains QDCOUNT (usually 1) entries, each of the following format:
//
type SectionQuestion struct {
	// A domain name represented as a sequence of labels, where each label
	// consists of a length octet followed by that number of octets.  The
	// domain name terminates with the zero length octet for the null
	// label of the root.  Note that this field may be an odd number of
	// octets; no padding is used.
	Name []byte

	// A two octet code which specifies the type of the query.  The values
	// for this field include all codes valid for a TYPE field, together
	// with some more general codes which can match more than one type of
	// RR.
	Type QueryType

	// A two octet code that specifies the class of the query.  For
	// example, the QCLASS field is IN for the Internet.
	Class uint16
}

//
// Reset the message question field to it's default values for query.
//
func (question *SectionQuestion) Reset() {
	question.Name = question.Name[:0]
	question.Type = QueryTypeA
	question.Class = QueryClassIN
}

//
// Size return the section question size, length of name + 2 (1 octet for
// beginning size plus 1 octer for end of label) + 2 octets of
// qtype + 2 octets of qclass
//
func (question *SectionQuestion) Size() int {
	return len(question.Name) + 6
}

//
// String will return the string representation of section question structure.
//
func (question *SectionQuestion) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "&{Name:%s Type:%d Class:%d}", question.Name, question.Type, question.Class)

	return buf.String()
}

//
// UnmarshalBinary unpack the DNS question section.
//
func (question *SectionQuestion) UnmarshalBinary(packet []byte) error {
	if len(packet) == 0 {
		return nil
	}

	count := packet[0]
	x := uint(1)

	for {
		for y := byte(0); y < count; y++ {
			question.Name = append(question.Name, packet[x])
			x++
		}
		count = packet[x]
		x++
		if count == 0 {
			break
		}
		question.Name = append(question.Name, '.')
	}

	question.Type = QueryType(libbytes.ReadUint16(packet, x))
	x += 2
	question.Class = uint16(libbytes.ReadUint16(packet, x))
	x += 2

	return nil
}
