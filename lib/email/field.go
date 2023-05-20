// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
	libtime "github.com/shuLhan/share/lib/time"
)

// Field represent field name and value in header.
type Field struct {
	// ContentType contains unpacked value of field with Name
	// "Content-Type" or nil if still packed.
	ContentType *ContentType

	date   *time.Time
	mboxes []*Mailbox

	// Name contains "relaxed" canonicalization of field name.
	Name []byte
	// Value contains "relaxed" canonicalization of field value.
	Value []byte

	// oriName contains "simple" canonicalization of field name.
	oriName []byte
	// oriValue contains "simple" canonicalization of field value.
	oriValue []byte

	// Type of field, the numeric representation of field name.
	Type FieldType

	// true if field.unpack has been called, false when field.setValue is
	// called again.
	unpacked bool
}

// ParseField create and initialize Field by parsing a single line message
// header field from raw input.
//
// If raw input contains multiple lines, the rest of lines will be returned.
//
// On error, it will return nil Field, and rest will contains the beginning of
// invalid input.
func ParseField(raw []byte) (field *Field, rest []byte, err error) {
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
			err = fmt.Errorf("email: invalid field at '%s'", raw[:x])
			goto invalid
		}
	}
	if len(raw) == x {
		err = fmt.Errorf("email: invalid field at '%s'", raw[:x])
		goto invalid
	}

	// Skip WSP before ':'.
	for ; x < len(raw) && (raw[x] == '\t' || raw[x] == ' '); x++ {
	}
	if len(raw) == x {
		err = fmt.Errorf("email: invalid field at '%s'", raw[:x])
		goto invalid
	}
	if raw[x] != ':' {
		err = fmt.Errorf("email: missing field separator at '%s'", raw[:x])
		goto invalid
	}

	field.setName(raw[:x])
	x++
	start = x

	// Skip WSP after ':'.
	for ; x < len(raw) && (raw[x] == '\t' || raw[x] == ' '); x++ {
	}

	if len(raw) == x {
		err = fmt.Errorf("email: empty field value at '%s'", raw[:x])
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
				err = fmt.Errorf("email: invalid field value at '%s'", raw[:x])
				goto invalid
			}
		}
		if x == len(raw) || raw[x] != lf {
			err = fmt.Errorf("email: field value without CRLF at '%s'", raw[:x])
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
		err = fmt.Errorf("email: field line greater than 998 characters")
		return nil, nil, err
	}

	field.setValue(raw[start:x])

	if len(field.Value) == 0 {
		err = fmt.Errorf("email: empty field value at '%s'", raw[:x])
		goto invalid
	}

	if len(raw) > x {
		rest = raw[x:]
	}

	return field, rest, nil

invalid:
	if x < len(raw) {
		rest = raw[x:]
	}
	return nil, rest, err
}

// addMailboxes append zero or more mailboxes to current mboxes.
func (field *Field) addMailboxes(mailboxes []byte) (err error) {
	var (
		mboxes []*Mailbox
	)

	mailboxes = bytes.TrimSpace(mailboxes)

	mboxes, err = ParseMailboxes(mailboxes)
	if err != nil {
		return err
	}
	field.mboxes = append(field.mboxes, mboxes...)

	if len(field.Value) > 0 {
		field.Value = bytes.TrimSpace(field.Value)
		field.Value = append(field.Value, ',', ' ')
	}
	field.appendValue(mailboxes)

	return nil
}

func (field *Field) appendValue(raw []byte) {
	var (
		x      int
		spaces int
	)

	// Skip leading spaces.
	for ; x < len(raw); x++ {
		if !ascii.IsSpace(raw[x]) {
			break
		}
	}

	for ; x < len(raw); x++ {
		if ascii.IsSpace(raw[x]) {
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

// setName set field Name by canonicalizing raw field name using "simple" and
// "relaxed" algorithms.
// .
// "simple" algorithm store raw field name as is.
//
// "relaxed" algorithm convert field name to lowercase and removing trailing
// whitespaces.
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

// setValue set the field Value by canonicalizing raw input using "simple" and
// "relaxed" algorithms.
//
// "simple" algorithm store raw field value as is in oriValue.
//
// "relaxed" algorithm remove leading and trailing WSP, replacing all
// CFWS with single space, but not removing CRLF at end.
func (field *Field) setValue(raw []byte) {
	field.oriValue = raw
	field.Value = make([]byte, 0, len(raw))
	field.appendValue(raw)
}

// Relaxed return the relaxed canonicalization of field name and value.
func (field *Field) Relaxed() (out []byte) {
	out = make([]byte, 0, len(field.Name)+len(field.Value)+1)
	out = append(out, field.Name...)
	out = append(out, ':')
	out = append(out, field.Value...)
	return
}

// Simple return the simple canonicalization of field name and value.
func (field *Field) Simple() (out []byte) {
	out = make([]byte, 0, len(field.oriName)+len(field.oriValue)+1)
	out = append(out, field.oriName...)
	out = append(out, ':')
	out = append(out, field.oriValue...)
	return
}

// unpack the field Value based on field Name.
func (field *Field) unpack() (err error) {
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

// updateType update the field type based on field name.
func (field *Field) updateType() {
	for k, v := range fieldNames {
		if bytes.Equal(v, field.Name) {
			field.Type = k
			return
		}
	}
	field.Type = FieldTypeOptional
}

// unpackDate from field value into time.Time.
//
// Format,
//
//	[day-of-week ","] day SP month SP year SP hour ":" minute [ ":" second ] SP zone
//
//	day-of-week = "Mon" / ... / "Sun"
//	day         = 1*2DIGIT
//	month       = "Jan" / ... / "Dec"
//	year        = 4*DIGIT
//	hour        = 2DIGIT
//	minute      = 2DIGIT
//	second      = 2DIGIT
//	zone        = ("+" / "-") 4DIGIT
func (field *Field) unpackDate() (err error) {
	var logp = `unpackDate`

	if len(field.Value) == 0 {
		return fmt.Errorf(`%s: empty date`, logp)
	}

	var (
		parser = libbytes.NewParser(field.Value, []byte{',', ' ', cr, lf})

		vstr  string
		token []byte
		c     byte
		ok    bool
	)

	token, c = parser.ReadNoSpace()
	parser.RemoveDelimiters([]byte{','})
	if c == ',' {
		var dow = string(token)
		for _, vstr = range libtime.ShortDayNames {
			if vstr == dow {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf(`%s: invalid day of week %s`, logp, dow)
		}
		token, _ = parser.ReadNoSpace()
	}

	// Get day ...
	var day int64
	day, err = strconv.ParseInt(string(token), 10, 64)
	if err != nil {
		return fmt.Errorf(`%s: invalid or missing day %s`, logp, token)
	}

	// Get month ...
	var month time.Month
	token, _ = parser.ReadNoSpace()
	month, ok = libtime.ShortMonths[string(token)]
	if !ok {
		return fmt.Errorf(`%s: invalid or missing month %s`, logp, token)
	}

	// Get year ...
	var year int64
	token, _ = parser.ReadNoSpace()
	year, err = strconv.ParseInt(string(token), 10, 64)
	if err != nil {
		return fmt.Errorf(`%s: invalid or missing year %s`, logp, token)
	}

	// Obsolete year format allow two or three digits.
	switch {
	case year < 50:
		year += 2000
	case year >= 50 && year < 1000:
		year += 1900
	}

	parser.AddDelimiters([]byte{':'})

	// Get hour ...
	var hour int64
	token, c = parser.ReadNoSpace()
	hour, err = strconv.ParseInt(string(token), 10, 64)
	if err != nil {
		return fmt.Errorf(`%s: invalid or missing hour %s`, logp, token)
	}
	if hour < 0 || hour > 23 {
		return fmt.Errorf(`%s: invalid hour %d`, logp, hour)
	}
	if c != ':' {
		return fmt.Errorf(`%s: invalid or missing time separator`, logp)
	}

	// Get minute ...
	var min int64
	token, c = parser.ReadNoSpace()
	min, err = strconv.ParseInt(string(token), 10, 64)
	if err != nil {
		return fmt.Errorf(`%s: invalid or missing minute %s`, logp, token)
	}
	if min < 0 || min > 59 {
		return fmt.Errorf(`%s: invalid minute %d`, logp, min)
	}

	// Get second ...
	var sec int64
	if c == ':' {
		parser.RemoveDelimiters([]byte{':'})
		token, _ = parser.ReadNoSpace()
		sec, err = strconv.ParseInt(string(token), 10, 64)
		if err != nil {
			return fmt.Errorf(`%s: invalid second %s`, logp, token)
		}
		if sec < 0 || sec > 59 {
			return fmt.Errorf(`%s: invalid second %d`, logp, sec)
		}
	}

	// Get zone offset ...
	var (
		off  int64
		zone string
	)
	token, _ = parser.ReadNoSpace()
	if len(token) == 0 {
		return fmt.Errorf(`%s: invalid or missing zone %s`, logp, token)
	}
	if len(token) != 0 {
		if token[0] == '+' || token[0] == '-' {
			off, err = strconv.ParseInt(string(token), 10, 64)
			if err != nil {
				return fmt.Errorf(`%s: invalid or missing zone offset %s`, logp, token)
			}
			zone = `UTC`
		} else {
			zone = string(token)
		}
	}

	var (
		loc = time.FixedZone(zone, computeOffSeconds(off))
		td  = time.Date(int(year), month, int(day), int(hour), int(min), int(sec), 0, loc)
	)

	field.date = &td
	field.unpacked = true

	return nil
}

func computeOffSeconds(off int64) int {
	hour := int(off / 100)
	min := int(off) - (hour * 100)
	return ((hour * 60) + min) * 60
}

// unpackMailboxList unpack list of mailbox from field Value.
func (field *Field) unpackMailboxList() (err error) {
	field.mboxes, err = ParseMailboxes(field.Value)
	if err == nil {
		field.unpacked = true
	}
	return err
}

// unpackMailbox unpack the raw addresses in field Value.
// It will return an error if address is invalid or contains multiple
// addresses.
func (field *Field) unpackMailbox() (err error) {
	mboxes, err := ParseMailboxes(field.Value)
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

// unpackContentType parse "Content-Type" from field Value.
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
