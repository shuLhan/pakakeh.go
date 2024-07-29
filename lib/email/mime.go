// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

// MIME represent part of message body with their header and content.
type MIME struct {
	contentType *ContentType

	Header  *Header
	Content []byte
}

// newMIME append new body with specific content type and charset.
// The content must be in raw format and it will be encoded using
// quoted-printable encoding.
func newMIME(contentType, content []byte) (mime *MIME, err error) {
	mime = &MIME{
		Header: &Header{},
	}

	err = mime.Header.Set(FieldTypeMIMEVersion, []byte(mimeVersion1))
	if err != nil {
		return nil, err
	}

	err = mime.Header.Set(FieldTypeContentType, contentType)
	if err != nil {
		return nil, err
	}

	mime.contentType = mime.Header.ContentType()

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

// ParseBodyPart parse one body part using boundary and return the rest of
// body.
func ParseBodyPart(raw, boundary []byte) (mime *MIME, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, raw, nil
	}
	if len(boundary) == 0 {
		return nil, raw, errors.New("ParseBodyPart: boundary parameter is empty")
	}

	var (
		parser = libbytes.NewParser(raw, []byte{lf})
		minlen = len(boundary) + 2

		line []byte
	)

	// find boundary ...
	parser.SkipSpaces()
	line, _ = parser.Read()
	rest, _ = parser.Stop()

	if len(line) == 0 {
		return nil, rest, nil
	}

	line = append(line, lf)
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
		return nil, rest, nil
	}

	mime = &MIME{}
	mime.Header, rest, err = ParseHeader(rest)
	if err != nil {
		return nil, raw, err
	}

	parser.Reset(rest, []byte{lf})

	for {
		line, _ = parser.Read()
		if len(line) == 0 {
			break
		}
		line = append(line, lf)
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
		parser.UnreadN(len(line))
		break
	}

	rest, _ = parser.Stop()

	return mime, rest, err
}

func (mime *MIME) decode(encoding string) (err error) {
	var logp = `decode`

	if mime.Header != nil {
		var partEncoding []*Field = mime.Header.Filter(FieldTypeContentTransferEncoding)
		var npart = len(partEncoding)
		if npart > 0 {
			encoding = strings.TrimSpace(partEncoding[npart-1].Value)
		}
	}

	switch encoding {
	case encodingBase64:
		err = mime.decodeBase64()
	case encodingQuotedPrintable:
		err = mime.decodeQuotedPrintable()
	default:
		// NO-OP.
	}
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

func (mime *MIME) decodeBase64() (err error) {
	var n = base64.RawStdEncoding.DecodedLen(len(mime.Content))
	var dest = make([]byte, n)
	n, err = base64.RawStdEncoding.Decode(dest, mime.Content)
	if err != nil {
		return fmt.Errorf(`decodeBase64: %w`, err)
	}
	mime.Content = dest[:n]
	return nil
}

func (mime *MIME) decodeQuotedPrintable() (err error) {
	var (
		logp = `decodeQuotedPrintable`
		qpr  = quotedprintable.NewReader(bytes.NewReader(mime.Content))
	)

	mime.Content, err = io.ReadAll(qpr)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

func (mime *MIME) isContentType(top, sub string) bool {
	if strings.EqualFold(mime.contentType.Top, top) {
		return strings.EqualFold(mime.contentType.Sub, sub)
	}
	return false
}

// String return string representation of MIME object.
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

// WriteTo write the MIME header and content into Writer w.
func (mime *MIME) WriteTo(w io.Writer) (n int64, err error) {
	var m int

	n, err = mime.Header.WriteTo(w)
	if err != nil {
		return n, err
	}

	m, err = w.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}
	n += int64(m)

	m, err = w.Write(mime.Content)
	if err != nil {
		return n, err
	}
	n += int64(m)

	return n, nil
}
