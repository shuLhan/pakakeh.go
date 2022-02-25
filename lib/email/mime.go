// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"errors"
	"mime/quotedprintable"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

//
// MIME represent part of message body with their header and content.
//
type MIME struct {
	contentType *ContentType

	Header  *Header
	Content []byte
}

//
// newMIME append new body with specific content type and charset.
// The content must be in raw format and it will be encoded using
// quoted-printable encoding.
//
func newMIME(contentType, content []byte) (mime *MIME, err error) {
	mime = &MIME{
		Header: &Header{},
	}

	err = mime.Header.Set(FieldTypeContentType, contentType)
	if err != nil {
		return nil, err
	}

	mime.contentType = mime.Header.ContentType()

	err = mime.Header.Set(FieldTypeMIMEVersion, []byte(mimeVersion1))
	if err != nil {
		return nil, err
	}

	err = mime.Header.Set(FieldTypeContentTransferEncoding, []byte(encodingQuotedPrintable))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	content = append(content, cr, lf)
	w := quotedprintable.NewWriter(&buf)
	_, err = w.Write(content)
	if err != nil {
		return nil, err
	}
	w.Close()

	mime.Content = buf.Bytes()

	return mime, nil
}

//
// ParseBodyPart parse one body part using boundary and return the rest of
// body.
//
func ParseBodyPart(raw, boundary []byte) (mime *MIME, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, raw, nil
	}
	if len(boundary) == 0 {
		return nil, raw, errors.New("ParseBodyPart: boundary parameter is empty")
	}

	r := &libio.Reader{}
	r.Init(raw)
	var (
		line   []byte
		minlen = len(boundary) + 2
	)

	// find boundary ...
	r.SkipSpaces()
	line = r.ReadLine()
	if len(line) == 0 {
		rest = r.Rest()
		return nil, rest, nil
	}
	if len(line) < minlen {
		return nil, raw, errors.New("ParseBodyPart: missing boundary line")
	}
	if line[len(line)-2] != cr {
		return nil, raw, errors.New("ParseBodyPart: invalid boundary line: missing CR")
	}
	if !bytes.Equal(line[:2], boundSeps) {
		return nil, raw, errors.New("ParseBodyPart: invalid boundary line: missing '--'")
	}
	if !bytes.Equal(line[2:minlen], boundary) {
		return nil, raw, errors.New("ParseBodyPart: boundary mismatch")
	}
	if bytes.Equal(line[minlen:len(line)-2], boundSeps) {
		// End of body.
		return nil, r.Rest(), nil
	}

	mime = &MIME{}
	mime.Header, rest, err = ParseHeader(r.Rest())
	if err != nil {
		return nil, raw, err
	}

	r.Init(rest)

	for {
		line = r.ReadLine()
		if len(line) == 0 {
			break
		}
		if len(line) < minlen {
			mime.Content = append(mime.Content, line...)
			continue
		}
		if line[len(line)-2] != cr {
			mime.Content = append(mime.Content, line...)
			continue
		}
		if !bytes.Equal(line[:2], boundSeps) {
			mime.Content = append(mime.Content, line...)
			continue
		}
		if !bytes.Equal(line[2:minlen], boundary) {
			mime.Content = append(mime.Content, line...)
			continue
		}
		r.UnreadN(len(line))
		break
	}

	rest = r.Rest()

	return mime, rest, err
}

//
// String return string representation of MIME object.
//
func (mime *MIME) String() string {
	var sb strings.Builder

	if mime.Header != nil {
		sb.Write(mime.Header.Relaxed())
	}
	sb.WriteByte(cr)
	sb.WriteByte(lf)
	sb.Write(mime.Content)

	return sb.String()
}
