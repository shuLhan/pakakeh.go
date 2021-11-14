// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libio "github.com/shuLhan/share/lib/io"
	libtime "github.com/shuLhan/share/lib/time"
)

const (
	defMinimumTTL = 3600
)

const (
	parseRRStart = 0
	parseRRTTL   = 1
	parseRRClass = 2
	parseRRType  = 4
)

const (
	parseSOAStart   = 0
	parseSOASerial  = 1
	parseSOARefresh = 2
	parseSOARetry   = 4
	parseSOAExpire  = 8
	parseSOAMinimum = 16
	parseSOAEnd     = 31
)

const (
	parseSRVService = 1 << iota
	parseSRVProto
	parseSRVName
	parseSRVPriority
	parseSRVWeight
	parseSRVPort
	parseSRVTarget
)

type zoneParser struct {
	zone   *ZoneFile
	lineno int
	seps   []byte
	terms  []byte
	reader *libio.Reader
	lastRR *ResourceRecord
	origin string
	ttl    uint32
	flag   int
}

func newZoneParser(file string) *zoneParser {
	return &zoneParser{
		zone:   NewZoneFile(file, ""),
		lineno: 1,
		seps:   []byte{' ', '\t'},
		terms:  []byte{';', '\n'},
	}
}

//
// Init parse zoneParser file from string.
//
func (m *zoneParser) Init(data, origin string, ttl uint32) {
	m.zone = NewZoneFile("(data)", "")
	m.lineno = 1
	m.origin = strings.ToLower(origin)
	m.ttl = ttl
	if m.reader == nil {
		m.reader = new(libio.Reader)
	}
	m.reader.Init([]byte(data))
}

//
// The format of these files is a sequence of entries.  Entries are
// predominantly line-oriented, though parentheses can be used to continue
// a list of items across a line boundary, and text literals can contain
// CRLF within the text.  Any combination of tabs and spaces act as a
// delimiter between the separate items that make up an entry.  The end of
// any line in the zone file can end with a comment.  The comment starts
// with a ";" (semicolon).
//
// The following entries are defined:
//
//    <blank>[<comment>]
//
//    $ORIGIN <domain-name> [<comment>]
//
//    $INCLUDE <file-name> [<domain-name>] [<comment>]
//
//    <domain-name><rr> [<comment>]
//
//    <blank><rr> [<comment>]
//
// Blank lines, with or without comments, are allowed anywhere in the file.
//
// Two control entries are defined: $ORIGIN and $INCLUDE.  $ORIGIN is
// followed by a domain name, and resets the current origin for relative
// domain names to the stated name.  $INCLUDE inserts the named file into
// the current file, and may optionally specify a domain name that sets the
// relative domain name origin for the included file.  $INCLUDE may also
// have a comment.  Note that a $INCLUDE entry never changes the relative
// origin of the parent file, regardless of changes to the relative origin
// made within the included file.
//
// The last two forms represent RRs.  If an entry for an RR begins with a
// blank, then the RR is assumed to be owned by the last stated owner.  If
// an RR entry begins with a <domain-name>, then the owner name is reset.
//
// <domain-name>s make up a large share of the data in the zone file.
// The labels in the domain name are expressed as character strings and
// separated by dots.  Quoting conventions allow arbitrary characters to be
// stored in domain names.  Domain names that end in a dot are called
// absolute, and are taken as complete.  Domain names which do not end in a
// dot are called relative; the actual domain name is the concatenation of
// the relative part with an origin specified in a $ORIGIN, $INCLUDE, or as
// an argument to the zone file loading routine.  A relative name is an
// error when no origin is available.
//
// <character-string> is expressed in one or two ways: as a contiguous set
// of characters without interior spaces, or as a string beginning with a "
// and ending with a ".  Inside a " delimited string any character can
// occur, except for a " itself, which must be quoted using \ (back slash).
//
// Because these files are text files several special encodings are
// necessary to allow arbitrary data to be loaded.  In particular:
//
// @               A free standing @ is used to denote the current origin.
//
// \X              where X is any character other than a digit (0-9), is
//                 used to quote that character so that its special meaning
//                 does not apply.  For example, "\." can be used to place
//                 a dot character in a label.
//
// \DDD            where each D is a digit is the octet corresponding to
//                 the decimal number described by DDD.  The resulting
//                 octet is assumed to be text and is not checked for
//                 special meaning.
//
// ( )             Parentheses are used to group data that crosses a line
//                 boundary.  In effect, line terminations are not
//                 recognized within parentheses.
//
// ;               Semicolon is used to start a comment; the remainder of
//                 the line is ignored.
//
func (m *zoneParser) parse() (err error) {
	var rr *ResourceRecord

	for {
		n, c := m.reader.SkipHorizontalSpace()
		if c == 0 {
			break
		}
		if c == '\n' || c == ';' {
			m.reader.SkipLine()
			m.lineno++
			continue
		}

		tok, isTerm, _ := m.reader.ReadUntil(m.seps, m.terms)
		if isTerm {
			return fmt.Errorf("line %d: invalid line %q",
				m.lineno, m.reader.Rest())
		}

		tok = ascii.ToUpper(tok)
		stok := string(tok)

		switch stok {
		case "$ORIGIN":
			err = m.parseDirectiveOrigin()
		case "$INCLUDE":
			err = m.parseDirectiveInclude()
		case "$TTL":
			err = m.parseDirectiveTTL()
		case "@":
			rr, err = m.parseRR(nil, tok)
		default:
			if n == 0 {
				rr, err = m.parseRR(nil, tok)
			} else {
				rr, err = m.parseRR(m.lastRR, tok)
			}
		}
		if err != nil {
			return err
		}
		if rr != nil {
			err = m.push(rr)
			if err != nil {
				return err
			}
		}
	}

	if m.ttl == 0 {
		m.ttl = defMinimumTTL
	}

	m.setMinimumTTL()
	m.pack()

	return nil
}

//
//    $ORIGIN <domain-name> [<comment>]
//
func (m *zoneParser) parseDirectiveOrigin() (err error) {
	_, c := m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: empty $origin directive", m.lineno)
	}

	tok, isTerm, c := m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: empty $origin directive", m.lineno)
	}

	m.origin = strings.TrimSuffix(strings.ToLower(string(tok)), ".")

	if isTerm {
		if c == ';' {
			m.reader.SkipLine()
		}
		m.lineno++
	} else {
		c = m.reader.SkipSpaces()
		if c == 0 {
			return nil
		}
		if c == ';' {
			m.reader.SkipLine()
			m.lineno++
		} else {
			return fmt.Errorf("line %d: invalid character '%c' after '%s'",
				m.lineno, c, tok)
		}
	}

	return nil
}

//
//    $INCLUDE <file-name> [<domain-name>] [<comment>]
//
func (m *zoneParser) parseDirectiveInclude() (err error) {
	_, c := m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: empty $include directive", m.lineno)
	}

	tok, isTerm, c := m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: empty $include directive", m.lineno)
	}

	var incfile, dname string

	tok = ascii.ToLower(tok)
	incfile = string(tok)

	// check if include followed by domain name.
	if !isTerm {
		c = m.reader.SkipSpaces()
	}

	if c == ';' {
		m.reader.SkipLine()
		m.lineno++
	} else if c != 0 {
		tok, isTerm, c = m.reader.ReadUntil(m.seps, m.terms)
		if !isTerm {
			c = m.reader.SkipSpaces()
		}
		if c != ';' {
			return fmt.Errorf("line %d: invalid character '%c' after '%s'",
				m.lineno, c, tok)
		}

		m.reader.SkipLine()
		m.lineno++

		if len(tok) > 0 {
			dname = string(tok)
		} else {
			dname = m.origin
		}
	}

	zoneFile, err := ParseZoneFile(incfile, dname, m.ttl)
	if err != nil {
		return err
	}

	m.zone.messages = append(m.zone.messages, zoneFile.messages...)

	return nil
}

func (m *zoneParser) parseDirectiveTTL() (err error) {
	var (
		c      byte
		isTerm bool
		tok    []byte
	)

	_, c = m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: empty $TTL directive", m.lineno)
	}

	tok, isTerm, _ = m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: empty $ttl directive", m.lineno)
	}

	tok = ascii.ToLower(tok)
	stok := string(tok)

	m.ttl, err = parseTTL(tok, stok)
	if err != nil {
		return err
	}

	if isTerm {
		m.reader.SkipLine()
		m.lineno++
	} else {
		c = m.reader.SkipSpaces()
		if c == 0 {
			return nil
		}
		if c == ';' {
			m.reader.SkipLine()
			m.lineno++
		} else {
			return fmt.Errorf("line %d: invalid character '%c' after '%s'",
				m.lineno, c, tok)
		}
	}

	return nil
}

//
// parseTTL convert characters of duration, with or without unit, to seconds.
//
func parseTTL(tok []byte, stok string) (seconds uint32, err error) {
	var (
		v   uint64
		dur time.Duration
	)

	if ascii.IsDigits(tok) {
		v, err = strconv.ParseUint(stok, 10, 32)
		if err != nil {
			return 0, err
		}

		seconds = uint32(v)

		return seconds, nil
	}

	dur, err = libtime.ParseDuration(stok)
	if err != nil {
		return 0, err
	}

	seconds = uint32(dur.Seconds())

	return seconds, nil
}

//
// <rr> contents take one of the following forms:
//
//    [<TTL>] [<class>] <type> <RDATA>
//
//    [<class>] [<TTL>] <type> <RDATA>
//
// The RR begins with optional TTL and class fields, followed by a type and
// RDATA field appropriate to the type and class.  Class and type use the
// standard mnemonics, TTL is a decimal integer.  Omitted class and TTL
// values are default to the last explicitly stated values.  Since type and
// class mnemonics are disjoint, the parse is unique.  (Note that this
// order is different from the order used in examples and the order used in
// the actual RRs; the given order allows easier parsing and defaulting.)
//
func (m *zoneParser) parseRR(prevRR *ResourceRecord, tok []byte) (
	rr *ResourceRecord, err error,
) {
	var (
		isTerm bool
	)

	stok := string(tok)

	rr = &ResourceRecord{}

	m.flag = 0

	if prevRR == nil {
		rr.Name = m.generateDomainName(tok)
		rr.TTL = m.ttl
		if m.lastRR != nil {
			rr.Class = m.lastRR.Class
		} else {
			rr.Class = RecordClassIN
		}
	} else {
		rr.Name = prevRR.Name
		rr.TTL = prevRR.TTL
		rr.Class = prevRR.Class

		if ascii.IsDigit(tok[0]) {
			ttl, err := parseTTL(tok, stok)
			if err != nil {
				return nil, err
			}
			rr.TTL = ttl
			m.flag |= parseRRTTL
		} else {
			ok := m.parseRRClassOrType(rr, stok)
			if !ok {
				err = fmt.Errorf("line %d: unknown class or type '%s'",
					m.lineno, stok)
				return nil, err
			}
		}
	}

	for {
		_, c := m.reader.SkipHorizontalSpace()
		if c == 0 || c == ';' {
			err = fmt.Errorf("line %d: invalid RR statement '%s'",
				m.lineno, stok)
			return nil, err
		}

		tok, isTerm, c = m.reader.ReadUntil(m.seps, m.terms)
		if len(tok) == 0 {
			err = fmt.Errorf("line %d: invalid RR statement '%s'",
				m.lineno, stok)
			return nil, err
		}

		orgtok := libbytes.Copy(tok)
		tok = ascii.ToUpper(tok)
		stok = string(tok)

		switch m.flag {
		case parseRRStart:
			if ascii.IsDigit(tok[0]) {
				rr.TTL, err = parseTTL(tok, stok)
				if err != nil {
					return nil, err
				}
				m.flag |= parseRRTTL
				continue
			}

			ok := m.parseRRClassOrType(rr, stok)
			if !ok {
				err = fmt.Errorf("line %d: unknown class or type '%s'", m.lineno, stok)
				return nil, err
			}

		case parseRRTTL:
			ok := m.parseRRClassOrType(rr, stok)
			if !ok {
				err = fmt.Errorf("line %d: unknown class or type '%s'", m.lineno, stok)
				return nil, err
			}

		case parseRRClass:
			if ascii.IsDigit(tok[0]) {
				rr.TTL, err = parseTTL(tok, stok)
				if err != nil {
					return nil, err
				}
				m.flag |= parseRRTTL
				continue
			}

			isType := m.parseRRType(rr, stok)
			if isType {
				m.flag |= parseRRType
				continue
			}

			err = fmt.Errorf("line %d: unknown type '%s'", m.lineno, stok)
			return nil, err

		case parseRRTTL | parseRRClass:
			isType := m.parseRRType(rr, stok)
			if isType {
				m.flag |= parseRRType
				continue
			}

			err = fmt.Errorf("line %d: unknown class or type '%s'", m.lineno, stok)
			return nil, err

		case parseRRType,
			parseRRTTL | parseRRType,
			parseRRClass | parseRRType,
			parseRRTTL | parseRRClass | parseRRType:

			if rr.Type == RecordTypeTXT {
				if !isTerm {
					orgtok = append(orgtok, c)
				}
			}

			err := m.parseRRData(rr, orgtok)
			if err != nil {
				return nil, err
			}
			goto out
		}
	}
out:
	return rr, nil
}

//
// parseRRClassOrType check if token either class or type.
// It will return true if one of them is set, otherwise it will return false.
//
func (m *zoneParser) parseRRClassOrType(rr *ResourceRecord, stok string) bool {
	isClass := m.parseRRClass(rr, stok)
	if isClass {
		m.flag |= parseRRClass
		return true
	}

	isType := m.parseRRType(rr, stok)
	if isType {
		m.flag |= parseRRType
		return true
	}

	return false
}

//
// parseRRClass check if token is known class.
// It will set the rr.Class and return true if stok is one of known class;
// otherwise it will return false.
//
func (m *zoneParser) parseRRClass(rr *ResourceRecord, stok string) bool {
	for k, v := range RecordClasses {
		if stok == k {
			rr.Class = v
			return true
		}
	}
	return false
}

//
// parseRRType check if token is one of known query type.
// It will set rr.Type and return true if token found, otherwise it will
// return false.
//
func (m *zoneParser) parseRRType(rr *ResourceRecord, stok string) bool {
	for k, v := range RecordTypes {
		if stok == k {
			rr.Type = v
			return true
		}
	}
	return false
}

func (m *zoneParser) parseRRData(rr *ResourceRecord, tok []byte) (err error) {
	switch rr.Type {
	case RecordTypeA, RecordTypeAAAA:
		rr.Value = string(tok)

	case RecordTypeNS, RecordTypeCNAME, RecordTypeMB, RecordTypeMG, RecordTypeMR, RecordTypePTR:
		rr.Value = m.generateDomainName(tok)

	case RecordTypeSOA:
		err = m.parseSOA(rr, tok)

	// NULL RRs are not allowed in zone files.
	case RecordTypeNULL:
		err = fmt.Errorf("line %d: NULL type is not allowed", m.lineno)

	// In zone files, both ports and protocols are expressed using
	// mnemonics or decimal numbers.
	case RecordTypeWKS:
		// TODO(ms)

	case RecordTypeHINFO:
		err = m.parseHInfo(rr, tok)

	case RecordTypeMINFO:
		err = m.parseMInfo(rr, tok)

	case RecordTypeMX:
		err = m.parseMX(rr, tok)

	case RecordTypeTXT:
		err = m.parseTXT(rr, tok)

	case RecordTypeSRV:
		err = m.parseSRV(rr, tok)
	}

	return err
}

func (m *zoneParser) parseSOA(rr *ResourceRecord, tok []byte) (err error) {
	tok = ascii.ToLower(tok)

	rrSOA := &RDataSOA{
		MName: m.generateDomainName(tok),
	}
	rr.Value = rrSOA

	_, c := m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: incomplete SOA values", m.lineno)
	}

	// Get RNAME
	tok, isTerm, _ := m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 || isTerm {
		return fmt.Errorf("line %d: Invalid SOA RNAME '%s'", m.lineno, tok)
	}

	tok = ascii.ToLower(tok)
	rrSOA.RName = m.generateDomainName(tok)

	var v int64
	isMultiline := false
	terms := []byte{'\n', ';'}

	// Get '(' or serial value
	tok, isTerm, _ = m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: invalid SOA SERIAL '%s'", m.lineno, tok)
	}

	if len(tok) == 1 && tok[0] == '(' {
		isMultiline = true
		terms = append(terms, ')')
		m.flag = parseSOAStart
	} else {
		v, err = strconv.ParseInt(string(tok), 10, 64)
		if err != nil {
			return err
		}
		rrSOA.Serial = uint32(v)
		m.flag = parseSOASerial
	}

	for {
		if isMultiline {
			c = m.reader.SkipSpaces()
			if c == ';' {
				m.reader.SkipLine()
				m.lineno++
				_ = m.reader.SkipSpaces()
			}
		} else {
			_, c = m.reader.SkipHorizontalSpace()
		}
		if c == 0 {
			break
		}

		tok, isTerm, c = m.reader.ReadUntil(m.seps, terms)
		if len(tok) == 0 {
			return fmt.Errorf("line %d: invalid SOA statement '%s'", m.lineno, tok)
		}
		if c == ';' {
			m.reader.SkipLine()
			m.lineno++
			_ = m.reader.SkipSpaces()
		}

		v, err = strconv.ParseInt(string(tok), 10, 64)
		if err != nil {
			return fmt.Errorf("line %d: invalid SOA value %s: %w",
				m.lineno, tok, err)
		}

		switch m.flag {
		case parseSOAStart:
			rrSOA.Serial = uint32(v)
			m.flag |= parseSOASerial

		case parseSOASerial:
			rrSOA.Refresh = int32(v)
			m.flag |= parseSOARefresh

		case parseSOASerial | parseSOARefresh:
			rrSOA.Retry = int32(v)
			m.flag |= parseSOARetry

		case parseSOASerial | parseSOARefresh | parseSOARetry:
			rrSOA.Expire = int32(v)
			m.flag |= parseSOAExpire

		case parseSOASerial | parseSOARefresh | parseSOARetry | parseSOAExpire:
			rrSOA.Minimum = uint32(v)
			m.flag |= parseSOAMinimum
			goto out

		default:
			return fmt.Errorf("line %d: invalid SOA flag %d '%s'",
				m.lineno, m.flag, string(tok))
		}
	}
	if m.flag != parseSOAEnd {
		return fmt.Errorf("line %d: incomplete SOA statement", m.lineno)
	}
out:
	if isMultiline {
		if isTerm {
			for c == ';' {
				m.reader.SkipLine()
				m.lineno++
				c = m.reader.SkipSpaces()
			}
			for c == '\n' {
				m.lineno++
				c = m.reader.SkipSpaces()
			}
		} else {
			c = m.reader.SkipSpaces()
		}

		if c != ')' {
			return fmt.Errorf("line %d: missing SOA closing parentheses", m.lineno)
		}

		_, _, c = m.reader.ReadUntil(m.seps, m.terms)
		if c == ';' {
			m.reader.SkipLine()
			m.lineno++
		}
	}

	if m.ttl == 0 {
		m.ttl = rrSOA.Minimum
	}

	return nil
}

func (m *zoneParser) parseHInfo(rr *ResourceRecord, tok []byte) (err error) {
	rrHInfo := &RDataHINFO{
		CPU: tok,
	}
	rr.Value = rrHInfo

	_, c := m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: missing HInfo OS value", m.lineno)
	}

	// Get OS
	tok, isTerm, _ := m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: missing HInfo OS value", m.lineno)
	}

	rrHInfo.OS = tok

	if !isTerm {
		m.reader.SkipLine()
		m.lineno++
	}

	return nil
}

func (m *zoneParser) parseMInfo(rr *ResourceRecord, tok []byte) (err error) {
	rrMInfo := &RDataMINFO{
		RMailBox: string(tok),
	}
	rr.Value = rrMInfo

	_, c := m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: missing MInfo EmailBox value", m.lineno)
	}

	// Get EmailBox value
	tok, isTerm, _ := m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: missing MInfo EmailBox value", m.lineno)
	}

	rrMInfo.EmailBox = string(tok)

	if !isTerm {
		m.reader.SkipLine()
		m.lineno++
	}

	return nil
}

func (m *zoneParser) parseMX(rr *ResourceRecord, tok []byte) (err error) {
	pref, err := strconv.ParseInt(string(tok), 10, 64)
	if err != nil {
		return fmt.Errorf("line %d: invalid MX Preference: %w", m.lineno, err)
	}

	rrMX := &RDataMX{
		Preference: int16(pref),
	}
	rr.Value = rrMX

	_, c := m.reader.SkipHorizontalSpace()
	if c == 0 || c == ';' {
		return fmt.Errorf("line %d: Missing MX Exchange value", m.lineno)
	}

	// Get EmailBox value
	tok, isTerm, _ := m.reader.ReadUntil(m.seps, m.terms)
	if len(tok) == 0 {
		return fmt.Errorf("line %d: missing MX Exchange value", m.lineno)
	}

	rrMX.Exchange = m.generateDomainName(tok)

	if !isTerm {
		m.reader.SkipLine()
		m.lineno++
	}

	return nil
}

//
// parseTXT parse TXT resource data.  The TXT rdata use the following format,
//
//	DQUOTE text DQUOTE
//
// The rdata MUST contains double quote at the beginning and end of text.
//
func (m *zoneParser) parseTXT(rr *ResourceRecord, v []byte) (err error) {
	tok, _, _ := m.reader.ReadUntil(nil, []byte{'\n'})
	v = append(v, tok...)
	v = bytes.TrimSpace(v)
	if v[0] != '"' {
		return fmt.Errorf("dns: missing start quote on TXT data")
	}
	if v[len(v)-1] != '"' {
		return fmt.Errorf("dns: missing end quote on TXT data")
	}
	v = v[1 : len(v)-1]

	rr.Value = string(v)

	return nil
}

func (m *zoneParser) parseSRV(rr *ResourceRecord, tok []byte) (err error) {
	var v int

	rrSRV := &RDataSRV{
		Service: string(tok),
	}
	rr.Value = rrSRV

	m.flag = parseSRVService

	for {
		_, c := m.reader.SkipHorizontalSpace()
		if c == 0 || c == ';' {
			return fmt.Errorf("line %d: incomplete SRV RDATA", m.lineno)
		}

		tok, _, _ = m.reader.ReadUntil(m.seps, m.terms)
		if len(tok) == 0 {
			return fmt.Errorf("line %d: incomplete SRV RDATA", m.lineno)
		}

		switch m.flag {
		case parseSRVService:
			rrSRV.Proto = string(tok)
			m.flag |= parseSRVProto

		case parseSRVService | parseSRVProto:
			rrSRV.Name = string(tok)
			m.flag |= parseSRVName

		case parseSRVService | parseSRVProto | parseSRVName:
			v, err = strconv.Atoi(string(tok))
			if err != nil {
				return fmt.Errorf("line %d: invalid SRV Priority value: %w",
					m.lineno, err)
			}
			rrSRV.Priority = uint16(v)
			m.flag |= parseSRVPriority

		case parseSRVService | parseSRVProto | parseSRVName | parseSRVPriority:
			v, err = strconv.Atoi(string(tok))
			if err != nil {
				return fmt.Errorf("line %d: invalid SRV Weight value: %w",
					m.lineno, err)
			}
			rrSRV.Weight = uint16(v)
			m.flag |= parseSRVWeight

		case parseSRVService | parseSRVProto | parseSRVName | parseSRVPriority | parseSRVWeight:
			v, err = strconv.Atoi(string(tok))
			if err != nil {
				return fmt.Errorf("line %d: invalid SRV Port value: %w",
					m.lineno, err)
			}
			rrSRV.Port = uint16(v)
			m.flag |= parseSRVPort

		case parseSRVService | parseSRVProto | parseSRVName | parseSRVPriority | parseSRVWeight | parseSRVPort:
			rrSRV.Target = string(tok)
			m.flag |= parseSRVTarget
			goto out

		default:
			return fmt.Errorf("line %d: invalid SRV RData", m.lineno)
		}
	}
out:
	_, c := m.reader.SkipHorizontalSpace()
	if c == ';' {
		m.reader.SkipLine()
		m.lineno++
	}

	return nil
}

func (m *zoneParser) generateDomainName(dname []byte) (out string) {
	dname = ascii.ToLower(dname)
	switch {
	case dname[0] == '@':
		out = m.origin
	case dname[len(dname)-1] == '.':
		out = string(dname[:len(dname)-1])
	default:
		out = string(dname) + "." + m.origin
	}
	return out
}

//
// push resource record (RR) into message answer only if domain name, type,
// and class already exist; otherwise it will create new message with question
// based on RR.
//
func (m *zoneParser) push(rr *ResourceRecord) error {
	m.lastRR = rr
	return m.zone.Add(rr)
}

func (m *zoneParser) setMinimumTTL() {
	for _, msg := range m.zone.messages {
		for x := 0; x < len(msg.Answer); x++ {
			if msg.Answer[x].TTL < m.ttl {
				msg.Answer[x].TTL = m.ttl
			}
		}
		for x := 0; x < len(msg.Authority); x++ {
			if msg.Authority[x].TTL < m.ttl {
				msg.Authority[x].TTL = m.ttl
			}
		}
		for x := 0; x < len(msg.Additional); x++ {
			if msg.Additional[x].TTL < m.ttl {
				msg.Additional[x].TTL = m.ttl
			}
		}
	}
}

func (m *zoneParser) pack() {
	for _, msg := range m.zone.messages {
		msg.Header.ANCount = uint16(len(msg.Answer))
		msg.Header.NSCount = uint16(len(msg.Authority))
		msg.Header.ARCount = uint16(len(msg.Additional))

		_, err := msg.Pack()
		if err != nil {
			log.Printf("! pack: %s\n", err)
			msg.Header.ANCount = 0
		}

		if debug.Value >= 3 {
			fmt.Printf("= Header: %+v\n", msg.Header)
			fmt.Printf("  Question: %s\n", msg.Question.String())
			for x := 0; x < len(msg.Answer); x++ {
				fmt.Printf("  Answer: %s\n", msg.Answer[x].String())
				fmt.Printf("  RData: %s\n", msg.Answer[x].Value)
			}
		}
	}
}
