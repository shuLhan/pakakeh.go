// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
)

// Header represent header of message.
// It contains list of field in message.
type Header struct {
	// Fields is ordered from top to bottom, the first field in message
	// header is equal to the first element in slice.
	//
	// We are not using map here it to prevent the header being reordered
	// when packing the message back into raw format.
	Fields []*Field
}

// ParseHeader parse the raw header from top to bottom.
//
// Raw header that start with CRLF indicate an empty header.
// In this case, it will return nil Header, indicating that no header was
// parsed, and the leading CRLF is removed on returned "rest".
//
// The raw header may end with optional CRLF, an empty line that separate
// header from body of message.
//
// On success it will return the rest of raw input (possible message's body)
// without leading CRLF.
func ParseHeader(raw []byte) (hdr *Header, rest []byte, err error) {
	var (
		field *Field
	)
	if len(raw) == 0 {
		return nil, nil, nil
	}

	rest = raw
	for len(rest) >= 2 {
		if rest[0] == lf {
			rest = rest[1:]
			return hdr, rest, nil
		}
		if rest[0] == cr && rest[1] == lf {
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
		hdr.Fields = append(hdr.Fields, field)
	}
	if len(rest) == 0 {
		return hdr, rest, nil
	}

	err = fmt.Errorf("ParseHeader: invalid end of header: '%s'", rest)

	return nil, rest, err
}

func (hdr *Header) addMailboxes(ft FieldType, mailboxes []byte) (err error) {
	var (
		f, field *Field
	)

	for _, f = range hdr.Fields {
		if f.Type == ft {
			field = f
			break
		}
	}
	if field == nil {
		field = &Field{
			Type: ft,
		}
		hdr.Fields = append(hdr.Fields, field)
		field.setName([]byte(fieldNames[ft]))
		field.setValue(mailboxes)
		return field.unpack()
	}

	return field.addMailboxes(mailboxes)
}

// Boundary return the message body boundary defined in Content-Type.
// If no field Content-Type or no boundary it will return nil.
func (hdr *Header) Boundary() string {
	ct := hdr.ContentType()
	if ct == nil {
		return ``
	}
	return ct.GetParamValue(ParamNameBoundary)
}

// ContentType return the unpacked value of field "Content-Type", or nil if no
// field Content-Type exist or there is an error when unpacking.
func (hdr *Header) ContentType() *ContentType {
	for _, f := range hdr.Fields {
		if f.Type != FieldTypeContentType {
			continue
		}
		if f.contentType == nil {
			err := f.unpack()
			if err != nil {
				log.Println("ContentType: ", err)
				return nil
			}
		}
		return f.contentType
	}
	return nil
}

// DKIM return sub-header of the "n" DKIM-Signature, start from the top.
// If no DKIM-Signature found it will return nil.
//
// For example, to get the second DKIM-Signature from the top, call it with
// "n=2", but if no second DKIM-Signature it will return nil.
func (hdr *Header) DKIM(n int) (dkimHeader *Header) {
	if n == 0 || len(hdr.Fields) == 0 {
		return nil
	}

	x := 0
	for ; x < len(hdr.Fields); x++ {
		if hdr.Fields[x].Type == FieldTypeDKIMSignature {
			n--
			if n == 0 {
				break
			}
		}
	}
	if x == len(hdr.Fields) {
		return nil
	}
	dkimHeader = &Header{
		Fields: make([]*Field, 0, len(hdr.Fields)-x),
	}
	for ; x < len(hdr.Fields); x++ {
		dkimHeader.Fields = append(dkimHeader.Fields, hdr.Fields[x])
	}

	return dkimHeader
}

// Filter specific field type.  If multiple fields type exist it will
// return all of them.
func (hdr *Header) Filter(ft FieldType) (fields []*Field) {
	for x := len(hdr.Fields) - 1; x >= 0; x-- {
		if hdr.Fields[x].Type == ft {
			fields = append(fields, hdr.Fields[x])
		}
	}
	return
}

// ID return the Message-ID or empty if not exist.
func (hdr *Header) ID() string {
	for x := len(hdr.Fields) - 1; x >= 0; x-- {
		if hdr.Fields[x].Type == FieldTypeMessageID {
			return hdr.Fields[x].oriValue
		}
	}
	return ``
}

// PushTop put the field at the top of header.
func (hdr *Header) PushTop(f *Field) {
	hdr.Fields = append([]*Field{f}, hdr.Fields...)
}

// Relaxed canonicalize the header using "relaxed" algorithm and return it.
func (hdr *Header) Relaxed() []byte {
	var bb bytes.Buffer

	for _, f := range hdr.Fields {
		if len(f.Name) > 0 && len(f.Value) > 0 {
			bb.Write(f.Relaxed())
		}
	}

	return bb.Bytes()
}

// Set the header's value based on specific type.
// If no field type found, the new field will be added to the list.
func (hdr *Header) Set(ft FieldType, value []byte) (err error) {
	var (
		field = &Field{
			Type: ft,
		}

		f *Field
		x int
	)

	field.setName([]byte(fieldNames[ft]))
	field.setValue(value)
	err = field.unpack()
	if err != nil {
		return fmt.Errorf("Set: %w", err)
	}

	for x, f = range hdr.Fields {
		if f.Type == ft {
			hdr.Fields[x] = field
			return nil
		}
	}
	hdr.Fields = append(hdr.Fields, field)
	return nil
}

// Simple canonicalize the header using "simple" algorithm.
func (hdr *Header) Simple() []byte {
	var bb bytes.Buffer

	for _, f := range hdr.Fields {
		if len(f.oriName) > 0 && len(f.oriValue) > 0 {
			bb.WriteString(f.oriName)
			bb.WriteByte(':')
			bb.WriteString(f.oriValue)
		}
	}

	return bb.Bytes()
}

// popByName remove the field where the name match from header.
func (hdr *Header) popByName(name string) (f *Field) {
	for x := len(hdr.Fields) - 1; x >= 0; x-- {
		if strings.EqualFold(hdr.Fields[x].Name, name) {
			f = hdr.Fields[x]
			hdr.Fields = append(hdr.Fields[:x], hdr.Fields[x+1:]...)
		}
	}
	return f
}

// SetMultipart make the header a multipart bodies with boundary.
func (hdr *Header) SetMultipart() (err error) {
	err = hdr.Set(FieldTypeMIMEVersion, []byte(mimeVersion1))
	if err != nil {
		return fmt.Errorf("email.SetMultipart: %w", err)
	}

	err = hdr.Set(FieldTypeContentType, []byte(contentTypeMultipartAlternative))
	if err != nil {
		return fmt.Errorf("email.SetMultipart: %w", err)
	}

	var boundary = randomString(32)
	contentType := hdr.ContentType()
	contentType.SetBoundary(boundary)

	return nil
}

// WriteTo the header into w.
// The header does not end with an empty line to allow multiple Header
// written multiple times.
func (hdr *Header) WriteTo(w io.Writer) (n int64, err error) {
	var (
		f *Field
		m int
	)
	for _, f = range hdr.Fields {
		switch f.Type {
		case FieldTypeContentType:
			m, err = fmt.Fprintf(w, "%s: %s\r\n", f.Name, f.contentType.String())
		case FieldTypeMessageID:
			m, err = fmt.Fprintf(w, "%s: <%s>\r\n", f.Name, f.oriValue)
		default:
			m, err = fmt.Fprintf(w, "%s: %s", f.Name, f.Value)
		}
		if err != nil {
			return n, err
		}
		n += int64(m)
	}
	return n, nil
}
