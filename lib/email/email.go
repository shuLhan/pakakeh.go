// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"strings"
)

var ( // nolint: gochecknoglobals
	crlf      = []byte{'\r', '\n'}
	boundSeps = []byte{'-', '-'}
)

//
// Message represent an unpacked internet message format.
//
type Message struct {
	Header  Header
	Body    Body
	oriBody []byte // oriBody contains original message body.
}

//
// ParseMessage parse the raw message header and body.
//
func ParseMessage(raw []byte) (msg *Message, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, nil, nil
	}

	msg = &Message{}

	rest, err = msg.Header.Unpack(raw)
	if err != nil {
		return nil, rest, err
	}

	msg.oriBody = rest
	boundary := msg.Header.Boundary()

	rest, err = msg.Body.Unpack(rest, boundary)

	return msg, rest, err
}

//
// String return the text representation of Message object.
//
func (msg *Message) String() string {
	var sb strings.Builder

	sb.WriteString(msg.Header.String())
	sb.WriteString("\r\n")
	sb.WriteString(msg.Body.String())

	return sb.String()
}
