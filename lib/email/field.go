// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"fmt"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	libio "github.com/shuLhan/share/lib/io"
	libtime "github.com/shuLhan/share/lib/time"
)

//
// Field represent field name and value in header.
//
type Field struct {
	// Type of field, the numeric representation of field name.
	Type FieldType
	// Name contains "relaxed" canonicalization of field name.
	Name []byte
	// Value contains "relaxed" canonicalization of field value.
	Value []byte

	// oriName contains "simple" canonicalization of field name.
	oriName []byte
	// oriValue contains "simple" canonicalization of field value.
	oriValue []byte

	date   *time.Time
	mboxes []*Mailbox

	// ContentType contains unpacked value of field with Name
	// "Content-Type" or nil if still packed.
	ContentType *ContentType

	// true if field.Unpack has been called, false when field.setValue is
	// called again.
	unpacked bool
}

//
// ParseField create and initialize Field by parsing a single line message
// header field from raw input.
//
// If raw input contains multiple lines, the rest of lines will be returned.
//
// On error, it will return nil Field, and rest will contains the beginning of
// invalid input.
//
func ParseField(raw []byte) (field *Field, rest []byte, err error) { // nolint: gocyclo
	if len(raw) == 0 {
		return nil, nil, nil
	}

	field = &Field{}
	isFolded := false
	start := 0

	// Get field's name.
	// Valid values: %d33-57 / %d59-126 .
	x := 0
	for ; x < len(raw); x++ {
		if raw[x] == ' ' || raw[x] == ':' {
			break
		}
		if raw[x] < 33 || raw[x] > 126 {
			goto invalid
		}
	}
	if len(raw) == x {
		goto invalid
	}

	// Skip spaces before ':'.
	for ; x < len(raw) && raw[x] == ' '; x++ {
	}
	if len(raw) == x {
		goto invalid
	}
	if raw[x] != ':' {
		goto invalid
	}

	field.setName(raw[:x])
	x++
	start = x

	// Skip WSP after ':'.
	for ; x < len(raw) && (raw[x] == '\t' || raw[x] == ' '); x++ {
	}
	if len(raw) == x {
		goto invalid
	}

	// Get field's value.
	// Valid values: WSP / %d33-126 .
	for ; x < len(raw); x++ {
		for ; x < len(raw); x++ {
			if raw[x] == '\t' || raw[x] == ' ' {
				continue
			}
			if raw[x] == cr {
				x++
				break
			}
			if raw[x] < 33 || raw[x] > 126 {
				goto invalid
			}
		}
		if x == len(raw) || raw[x] != lf {
			goto invalid
		}
		if x++; x == len(raw) {
			break
		}

		// Unfolding ...
		if raw[x] == '\t' || raw[x] == ' ' {
			isFolded = true
			continue
		}
		break
	}
	if !isFolded && x > 1000 {
		err = fmt.Errorf("ParseField: line greater than 998 characters")
		return nil, nil, err
	}

	field.setValue(raw[start:x])

	if len(field.Value) == 0 {
		goto invalid
	}

	if len(raw) > x {
		rest = raw[x:]
	}

	return field, rest, nil

invalid:
	if x < len(raw) {
		err = fmt.Errorf("ParseField: invalid character at index %d", x)
		rest = raw[x:]
	} else {
		err = fmt.Errorf("ParseField: invalid input")
	}
	return nil, rest, err
}

//
// setName set field Name by canonicalizing raw field name using "simple" and
// "relaxed" algorithms.
//.
// "simple" algorithm store raw field name as is.
//
// "relaxed" algorithm convert field name to lowercase and removing trailing
// whitespaces.
//
func (field *Field) setName(raw []byte) {
	field.oriName = raw
	field.Name = make([]byte, 0, len(raw))
	for x := 0; x < len(raw); x++ {
		if raw[x] == ' ' || raw[x] < 33 || raw[x] > 126 {
			break
		}
		if raw[x] >= 'A' && raw[x] <= 'Z' {
			field.Name = append(field.Name, raw[x]+32)
		} else {
			field.Name = append(field.Name, raw[x])
		}
	}
	field.updateType()
}

//
// setValue set the field Value by canonicalizing raw input using "simple" and
// "relaxed" algorithms.
//
// "simple" algorithm store raw field value as is in oriValue.
//
// "relaxed" algorithm remove leading and trailing WSP, replacing all
// CFWS with single space, but not removing CRLF at end.
//
func (field *Field) setValue(raw []byte) {
	field.oriValue = raw
	field.Value = make([]byte, 0, len(raw))

	x := 0
	// Skip leading spaces.
	for ; x < len(raw); x++ {
		if !libbytes.IsSpace(raw[x]) {
			break
		}
	}

	spaces := 0
	for ; x < len(raw); x++ {
		if libbytes.IsSpace(raw[x]) {
			spaces++
			continue
		}
		if spaces > 0 {
			field.Value = append(field.Value, ' ')
			spaces = 0
		}
		field.Value = append(field.Value, raw[x])
	}
	if len(field.Value) > 0 {
		field.Value = append(field.Value, cr)
		field.Value = append(field.Value, lf)
	}
	field.unpacked = false
}

//
// String return the relaxed canonicalization of field name and value
// separated by colon.
//
func (field *Field) String() string {
	return string(field.Name) + ":" + string(field.Value)
}

//
// Unpack the field Value based on field Name.
//
func (field *Field) Unpack() (err error) {
	switch field.Type {
	case FieldTypeDate:
		err = field.unpackDate()

	case FieldTypeFrom:
		err = field.unpackMailboxList()
	case FieldTypeSender:
		err = field.unpackMailbox()
	case FieldTypeReplyTo:
		err = field.unpackMailboxList()

	case FieldTypeTo:
		err = field.unpackMailboxList()
	case FieldTypeCC:
		err = field.unpackMailboxList()
	case FieldTypeBCC:
		err = field.unpackMailboxList()

	case FieldTypeResentDate:
		err = field.unpackDate()
	case FieldTypeResentFrom:
		err = field.unpackMailboxList()
	case FieldTypeResentSender:
		err = field.unpackMailbox()
	case FieldTypeResentTo:
		err = field.unpackMailboxList()
	case FieldTypeResentCC:
		err = field.unpackMailboxList()
	case FieldTypeResentBCC:
		err = field.unpackMailboxList()

	case FieldTypeReturnPath:
		err = field.unpackMailbox()

	case FieldTypeContentType:
		err = field.unpackContentType()
	}

	return err
}

//
// updateType update the field type based on field name.
//
func (field *Field) updateType() {
	for k, v := range fieldNames {
		if bytes.Equal(v, field.Name) {
			field.Type = k
			return
		}
	}
	field.Type = FieldTypeOptional
}

//
// unpackDate from field value into time.Time.
//
// Format,
//
//	[day-of-week ","] day month year hour ":" minute [ ":" second ] zone
//
//      day-of-week = "Mon" / ... / "Sun"
//      day         = 1*2DIGIT
//      month       = "Jan" / ... / "Dec"
//      year        = 4*DIGIT
//      hour        = 2DIGIT
//      minute      = 2DIGIT
//      second      = 2DIGIT
//	zone        = ("+" / "-") 4DIGIT
//
func (field *Field) unpackDate() (err error) {
	var (
		v              []byte
		ok             bool
		c              byte
		space          = []byte{' ', cr, lf}
		day, year      int64
		hour, min, sec int64
		off            int64
		month          time.Month
		loc            *time.Location
	)

	if len(field.Value) == 0 {
		return fmt.Errorf("unpackDate: empty date")
	}

	r := &libio.Reader{}
	r.Init(field.Value)

	c = r.SkipSpace()
	if !libbytes.IsDigit(c) {
		v, _, c = r.ReadUntil([]byte{','}, nil)
		if len(v) == 0 || c != ',' {
			return fmt.Errorf("unpackDate: invalid date format")
		}
		if c = r.SkipSpace(); c == 0 {
			return fmt.Errorf("unpackDate: invalid date format")
		}
	}

	// Get day ....
	if day, c = r.ScanInt64(); c == 0 || c != ' ' {
		return fmt.Errorf("unpackDate: missing month")
	}
	// Get month ...
	r.SkipSpace()
	v, _, _ = r.ReadUntil(space, nil)
	month, ok = libtime.ShortMonths[string(v)]
	if !ok {
		return fmt.Errorf("unpackDate: invalid month: '%s'", v)
	}

	// Get year ...
	r.SkipSpace()
	if year, c = r.ScanInt64(); c == 0 || c != ' ' {
		return fmt.Errorf("unpackDate: invalid year")
	}

	// Obsolete year allow two or three digits.
	switch {
	case year < 50:
		year += 2000
	case year >= 50 && year < 1000:
		year += 1900
	}

	// Get hour ...
	if hour, c = r.ScanInt64(); c == 0 || c != ':' {
		return fmt.Errorf("unpackDate: invalid hour")
	}
	if hour < 0 || hour > 23 {
		return fmt.Errorf("unpackDate: invalid hour: %d", hour)
	}

	// Get minute ...
	r.SkipN(1)
	min, c = r.ScanInt64()
	if min < 0 || min > 59 {
		return fmt.Errorf("unpackDate: invalid minute: %d", min)
	}

	// Get second ...
	if c == ':' {
		r.SkipN(1)
		sec, _ = r.ScanInt64()
		if sec < 0 || sec > 59 {
			return fmt.Errorf("unpackDate: invalid second: %d", sec)
		}
	}

	// Get zone offset ...
	c = r.SkipSpace()
	if c == 0 {
		return fmt.Errorf("unpackDate: missing zone")
	}
	off, _ = r.ScanInt64()

	loc = time.FixedZone("UTC", computeOffSeconds(off))
	td := time.Date(int(year), month, int(day), int(hour), int(min), int(sec), 0, loc)
	field.date = &td
	field.unpacked = true

	return err
}

func computeOffSeconds(off int64) int {
	hour := int(off / 100)
	min := int(off) - (hour * 100)
	return ((hour * 60) + min) * 60
}

//
// unpackMailboxList unpack list of mailbox from field Value.
//
func (field *Field) unpackMailboxList() (err error) {
	field.mboxes, err = ParseAddress(field.Value)
	if err == nil {
		field.unpacked = true
	}
	return err
}

//
// unpackMailbox unpack the raw addresses in field Value.
// It will return an error if address is invalid or contains multiple
// addresses.
//
func (field *Field) unpackMailbox() (err error) {
	mboxes, err := ParseAddress(field.Value)
	if err != nil {
		return err
	}
	if len(mboxes) != 1 {
		return fmt.Errorf("multiple address in %s: '%s'", field.Name,
			field.Value)
	}

	field.unpacked = true

	return nil
}

//
// unpackContentType parse "Content-Type" from field Value.
//
func (field *Field) unpackContentType() (err error) {
	if field.unpacked {
		return nil
	}

	ContentType, err := ParseContentType(field.Value)
	if err != nil {
		return err
	}

	field.ContentType = ContentType
	field.unpacked = true

	return nil
}
