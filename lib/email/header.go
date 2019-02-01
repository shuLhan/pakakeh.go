// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"fmt"
	"strings"
)

//
// Header represent list of field.
//
// We are not using map here it to prevent the header being reordeded when
// packing the message back into raw format.
//
type Header struct {
	fields []*Field
}

//
// Unpack the raw header from top to bottom.
//
// The raw header may end with optional CRLF, an empty line that separate
// header from body of message.
//
// On success it will return the rest of raw input (possible message's body)
// without leading CRLF.
//
func (hdr *Header) Unpack(raw []byte) ([]byte, error) {
	var (
		field *Field
		err   error
	)

	for len(raw) > 2 {
		field, raw, err = ParseField(raw)
		if err != nil {
			return raw, err
		}
		hdr.fields = append(hdr.fields, field)
		if len(raw) > 2 {
			if raw[0] == crlf[0] && raw[1] == crlf[1] {
				break
			}
		}
	}

	switch len(raw) {
	case 0:
	case 1:
		err = fmt.Errorf("Header.Unpack: invalid end of header: '%s'", raw)
	case 2:
		if raw[0] != crlf[0] || raw[1] != crlf[1] {
			err = fmt.Errorf("Header.Unpack: invalid end of header: '%s'", raw)
		} else {
			raw = raw[2:]
		}
	default:
		raw = raw[2:]
	}

	return raw, err
}

//
// String return the text representation of header, by concatenating all
// sanitized fields with CRLF.
//
func (hdr *Header) String() string {
	var sb strings.Builder

	for _, f := range hdr.fields {
		sb.WriteString(f.String())
	}

	return sb.String()
}
