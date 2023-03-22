// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"strings"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

// MessageQuestion contains the "question" in most queries.
type MessageQuestion struct {
	// The domain name to be queried.
	Name string

	// The Type of query.
	Type RecordType

	// The Class of the query.
	Class RecordClass
}

// Reset the message question field to it's default values for query.
func (qst *MessageQuestion) Reset() {
	qst.Name = ""
	qst.Type = RecordTypeA
	qst.Class = RecordClassIN
}

func (qst *MessageQuestion) String() string {
	return fmt.Sprintf("{Name:%s Type:%s}", qst.Name, RecordTypeNames[qst.Type])
}

// size return the section question size.
// The size depends on the Name.
// If the Name end with '.', it will return length of Name + 4 + 1; otherwise
// it will return length of Name + 4 + 2.
// The 4 is size of type and class, 1 is for the first length, and another 1
// for zero length at the end.
func (qst *MessageQuestion) size() int {
	var (
		size  = len(qst.Name)
		lastc = size - 1
	)
	if lastc >= 0 && qst.Name[lastc] == '.' {
		return size + 5
	}
	return size + 6
}

// unpack the DNS question section from packet.
func (qst *MessageQuestion) unpack(packet []byte) (err error) {
	if len(packet) == 0 {
		return nil
	}

	var (
		logp  = "MessageQuestion.unpack"
		sb    strings.Builder
		x     int
		y     int
		count int
	)

	for {
		count = int(packet[x])
		if count == 0 {
			x++
			break
		}
		if x+count+1 >= len(packet) {
			return fmt.Errorf("%s: label length overflow at index %d", logp, x)
		}
		if sb.Len() > 0 {
			sb.WriteByte('.')
		}
		for y = 0; y < count; y++ {
			x++
			if packet[x] >= 'A' && packet[x] <= 'Z' {
				sb.WriteByte(packet[x] + 32)
			} else {
				sb.WriteByte(packet[x])
			}
		}
		x++
	}

	if x+4 > len(packet) {
		return fmt.Errorf("%s: packet too small, missing type and/or class", logp)
	}

	qst.Name = sb.String()
	qst.Type = RecordType(libbytes.ReadUint16(packet, uint(x)))
	x += 2
	qst.Class = RecordClass(libbytes.ReadUint16(packet, uint(x)))

	return nil
}
