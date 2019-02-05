// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"fmt"
	"log"
	"strings"
)

//
// Header represent list of field.
//
//
type Header struct {
	// We are not using map here it to prevent the header being reordeded when
	// packing the message back into raw format.
	fields []*Field
}

//
// ParseHeader parse the raw header from top to bottom.
//
// Raw header that start with CRLF indicate an empty header.
// In this case, it will return nil Header, indicating that no header was
// parsed, and remove the leading CRLF on returned "rest".
//
// The raw header may end with optional CRLF, an empty line that separate
// header from body of message.
//
// On success it will return the rest of raw input (possible message's body)
// without leading CRLF.
//
func ParseHeader(raw []byte) (hdr *Header, rest []byte, err error) {
	var (
		field *Field
	)
	if len(raw) == 0 {
		return nil, nil, nil
	}

	rest = raw
	for len(rest) >= 2 {
		if rest[0] == '\r' && rest[1] == '\n' {
			rest = rest[2:]
			return hdr, rest, nil
		}

		field, rest, err = ParseField(rest)
		if err != nil {
			return nil, rest, err
		}
		if hdr == nil {
			hdr = &Header{}
		}
		hdr.fields = append(hdr.fields, field)
	}
	if len(rest) == 0 {
		return hdr, rest, nil
	}

	err = fmt.Errorf("ParseHeader: invalid end of header: '%s'", rest)

	return nil, rest, err
}

//
// Boundary return the message body boundary defined in Content-Type.
// If no field Content-Type or no boundary it will return nil.
//
func (hdr *Header) Boundary() []byte {
	ct := hdr.ContentType()
	if ct == nil {
		return nil
	}
	return ct.GetParamValue(ParamNameBoundary)
}

//
// ContentType return the unpacked value of field "Content-Type", or nil if no
// field Content-Type exist or there is an error when unpacking.
//
func (hdr *Header) ContentType() *ContentType {
	for _, f := range hdr.fields {
		if f.Type != FieldTypeContentType {
			continue
		}
		if f.ContentType == nil {
			err := f.Unpack()
			if err != nil {
				log.Println("ContentType: ", err)
				return nil
			}
		}
		return f.ContentType
	}
	return nil
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
