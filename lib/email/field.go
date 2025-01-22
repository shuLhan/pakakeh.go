// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package email

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	libtime "git.sr.ht/~shulhan/pakakeh.go/lib/time"
)

// Field represent field name and value in header.
type Field struct {
	// Name contains "relaxed" canonicalization of field name.
	Name string
	// Value contains "relaxed" canonicalization of field value.
	Value string

	// oriName contains "simple" canonicalization of field name.
	oriName string
	// oriValue contains "simple" canonicalization of field value.
	oriValue string

	// contentType contains unpacked value of field with Name
	// "Content-Type" or nil if still packed.
	contentType *ContentType

	date   *time.Time
	mboxes []*Mailbox

	// Params contains unpacked parameter from Value.
	// Not all content type has parameters.
	Params []Param

	// Type of field, the numeric representation of field name.
	Type FieldType

	// isFolded set to true if field line contains folding, CRLF
	// following by space and values.
	isFolded bool

	// true if field.unpack has been called, false when field.setValue
	// is called again.
	unpacked bool
}

// ParseField create and initialize Field by parsing a single line message
// header.
//
// If raw input contains multiple lines, the rest of lines will be returned.
//
// On error, it will return nil Field, and rest will contains the beginning
// of invalid input.
func ParseField(raw []byte) (field *Field, rest []byte, err error) {
	if len(raw) == 0 {
		return nil, nil, nil
	}

	var logp = `ParseField`

	field = &Field{}

	raw, err = field.parseName(raw)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	raw, err = field.parseValue(raw)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	if !field.isFolded {
		if (len(field.oriName) + len(field.oriValue) + 1) > 1000 {
			return nil, nil, fmt.Errorf(`%s: field line greater than 998 characters`, logp)
		}
	}

	field.updateType()
	err = field.unpack()
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	rest = raw
	return field, rest, nil
}

// parseName parse the field Name.
// Format,
//
//	field-name = 1*(ftext / obs-ftext) ":"
//	obs-ftext  = %d32 / ftext
//	           ; space allowed in [obsolete] specification.
//	[ftext]    = %d33-57 / %d59-126
//	           ; printable ASCII character except colon (%d58).
//
// [ftext]: https://datatracker.ietf.org/doc/html/rfc5322#section-2.2
// [obsolete]: https://datatracker.ietf.org/doc/html/rfc5322#section-4.5
func (field *Field) parseName(raw []byte) (rest []byte, err error) {
	var (
		logp = `parseName`
		x    int
	)
	for ; x < len(raw); x++ {
		if raw[x] == '\t' || raw[x] == ' ' || raw[x] == ':' {
			break
		}
		if raw[x] < 33 || raw[x] > 126 {
			return nil, fmt.Errorf(`%s: %q invalid character %q`, logp, raw[:x], raw[x])
		}
	}
	// Skip WSP before ':'.
	for x < len(raw) && (raw[x] == '\t' || raw[x] == ' ') {
		x++
	}
	if len(raw) == x {
		return nil, fmt.Errorf(`%s: missing value`, logp)
	}
	if raw[x] != ':' {
		return nil, fmt.Errorf(`%s: missing field separator`, logp)
	}

	field.setName(raw[:x])

	rest = raw[x+1:]

	return rest, nil
}

// parseValue parse field value.
// Format,
//
//	field-body = 1*(FWS / WSP / %d33-126) CRLF
//	FWS        = CRLF WSP              ; \r\n followed by space.
//	WSP        = %d9 / %d32            ; tab or space.
//
// [Reference]: https://datatracker.ietf.org/doc/html/rfc5322#section-2.2
func (field *Field) parseValue(raw []byte) (rest []byte, err error) {
	var (
		logp = `parseValue`
		x    int
	)

	for ; x < len(raw); x++ {
		for ; x < len(raw); x++ {
			if raw[x] == '\t' || raw[x] == ' ' {
				continue
			}
			if raw[x] == cr {
				x++
				break
			}
			if raw[x] == lf {
				break
			}
			if raw[x] < 33 || raw[x] > 126 {
				return nil, fmt.Errorf(`%s: invalid field value %q`, logp, raw[x])
			}
		}
		if x == len(raw) || raw[x] != lf {
			return nil, fmt.Errorf(`%s: invalid or missing termination`, logp)
		}
		x++
		if x == len(raw) {
			break
		}
		// Unfolding ...
		if raw[x] == '\t' || raw[x] == ' ' {
			field.isFolded = true
			continue
		}
		// End with CRLF.
		break
	}

	field.setValue(raw[:x])

	if len(field.Value) == 0 {
		return nil, fmt.Errorf(`%s: empty field value`, logp)
	}

	rest = raw[x:]

	return rest, nil
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
		field.Value = strings.TrimSpace(field.Value)
		field.Value += `, `
	}

	mailboxes = relaxedValue(mailboxes)
	field.Value += string(mailboxes)

	return nil
}

func relaxedName(raw []byte) (rel []byte) {
	rel = make([]byte, 0, len(raw))
	for x := range len(raw) {
		if raw[x] == ' ' || raw[x] < 33 || raw[x] > 126 {
			break
		}
		if raw[x] >= 'A' && raw[x] <= 'Z' {
			rel = append(rel, raw[x]+32)
		} else {
			rel = append(rel, raw[x])
		}
	}
	return rel
}

func relaxedValue(raw []byte) (rel []byte) {
	var (
		bb     bytes.Buffer
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
			bb.WriteByte(' ')
			spaces = 0
		}
		bb.WriteByte(raw[x])
	}
	if bb.Len() > 0 {
		bb.WriteByte(cr)
		bb.WriteByte(lf)
	}
	return bb.Bytes()
}

// setName set field Name by canonicalizing raw field name using "simple" and
// "relaxed" algorithms.
//
// "simple" algorithm store raw field name as is.
//
// "relaxed" algorithm convert field name to lowercase and removing trailing
// whitespaces.
func (field *Field) setName(raw []byte) {
	field.oriName = string(raw)

	raw = relaxedName(raw)
	field.Name = string(raw)
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
	field.oriValue = string(raw)

	raw = relaxedValue(raw)
	field.Value = string(raw)
	field.unpacked = false
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
	out = append(out, []byte(field.oriName)...)
	out = append(out, ':')
	out = append(out, []byte(field.oriValue)...)
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
		if strings.EqualFold(v, field.Name) {
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
		value  = sanitize([]byte(field.Value))
		parser = libbytes.NewParser(value, []byte{',', ' '})

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
	if c == ' ' {
		_, c = parser.SkipSpaces()
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
	token = nil

	if c == ' ' {
		token, c = parser.ReadNoSpace()
		if c == ':' && len(token) != 0 {
			return fmt.Errorf(`%s: unknown token after minute %q`, logp, token)
		}
		// At this point the date may have second and token may be a
		// zone.
		// We check again later if token is nil after parsing the
		// second part.
	}

	parser.RemoveDelimiters([]byte{':'})

	// Get second ...
	var sec int64
	if c == ':' {
		token, _ = parser.ReadNoSpace()
		sec, err = strconv.ParseInt(string(token), 10, 64)
		if err != nil {
			return fmt.Errorf(`%s: invalid second %s`, logp, token)
		}
		if sec < 0 || sec > 59 {
			return fmt.Errorf(`%s: invalid second %d`, logp, sec)
		}
		token = nil
	}

	// Get zone offset.
	var (
		off  int64
		zone string
	)
	if token == nil { // The data contains second.
		token, _ = parser.ReadNoSpace()
		if len(token) == 0 {
			return fmt.Errorf(`%s: invalid or missing zone %s`, logp, token)
		}
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
	field.mboxes, err = ParseMailboxes([]byte(field.Value))
	if err == nil {
		field.unpacked = true
	}
	return err
}

// unpackMailbox unpack the raw addresses in field Value.
// It will return an error if address is invalid or contains multiple
// addresses.
func (field *Field) unpackMailbox() (err error) {
	mboxes, err := ParseMailboxes([]byte(field.Value))
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

	field.contentType, err = ParseContentType([]byte(field.Value))
	if err != nil {
		return err
	}

	field.Params = field.contentType.Params
	field.unpacked = true

	return nil
}
