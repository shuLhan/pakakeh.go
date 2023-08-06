// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
	libtime "github.com/shuLhan/share/lib/time"
)

// List of flag for parsing RR.
const (
	flagRRStart = 0
	flagRRTtl   = 1
	flagRRClass = 2
	flagRRType  = 4 // Once the type has known the order is linear.
	flagRREnd   = 8
)

type zoneParser struct {
	zone   *Zone
	parser *libbytes.Parser
	lastRR *ResourceRecord
	lineno int
}

func newZoneParser(data []byte, zone *Zone) (zp *zoneParser) {
	zp = &zoneParser{}
	zp.Reset(data, zone)
	return zp
}

// Reset zoneParser by parsing from slice of byte.
func (m *zoneParser) Reset(data []byte, zone *Zone) {
	if zone == nil {
		zone = NewZone(`(data)`, ``)
	}

	m.zone = zone
	m.lineno = 1
	if m.parser == nil {
		m.parser = libbytes.NewParser(nil, nil)
	}

	data = bytes.TrimSpace(data)
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	m.parser.Reset(data, []byte{' ', '\t', '\n', ';'})
}

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
//	<blank>[<comment>]
//
//	$ORIGIN <domain-name> [<comment>]
//
//	$INCLUDE <file-name> [<domain-name>] [<comment>]
//
//	<domain-name><rr> [<comment>]
//
//	<blank><rr> [<comment>]
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
//	@ - A free standing @ is used to denote the current origin.
//
//	\X - where X is any character other than a digit (0-9), is used to
//	quote that character so that its special meaning does not apply.
//	For example, "\." can be used to place a dot character in a label.
//
//	\DDD - where each D is a digit is the octet corresponding to the
//	decimal number described by DDD.
//	The resulting octet is assumed to be text and is not checked for
//	special meaning.
//
//	( ) - Parentheses are used to group data that crosses a line
//	boundary.  In effect, line terminations are not	recognized within
//	parentheses.
//
//	; - Semicolon is used to start a comment; the remainder of the line is
//	ignored.
func (m *zoneParser) parse() (err error) {
	var (
		logp = `parse`

		rr   *ResourceRecord
		tok  []byte
		stok string
		n    int
		c    byte
	)

	for {
		// Check if the RR start with space or not.
		n, c = m.parser.SkipHorizontalSpaces()
		if c == 0 {
			break
		}
		if c == '\n' || c == ';' {
			m.parser.SkipLine()
			m.lineno++
			continue
		}

		tok, c = m.parser.ReadNoSpace()

		tok = ascii.ToUpper(tok)
		stok = string(tok)

		switch stok {
		case `$ORIGIN`:
			err = m.parseDirectiveOrigin(c)
		case `$INCLUDE`:
			err = m.parseDirectiveInclude(c)
		case `$TTL`:
			err = m.parseDirectiveTTL(c)
		case `@`:
			rr, err = m.parseRR(nil, tok)
		default:
			if n == 0 {
				rr, err = m.parseRR(nil, tok)
			} else {
				rr, err = m.parseRR(m.lastRR, tok)
			}
		}
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		if rr != nil {
			err = m.push(rr)
			if err != nil {
				return fmt.Errorf(`%s: %w`, logp, err)
			}
		}
	}

	m.setMinimumTTL()
	m.pack()

	return nil
}

// parseDirectiveOrigin parse the $ORIGIN directive in the following format,
//
//	$ORIGIN <domain-name> [<comment>]
func (m *zoneParser) parseDirectiveOrigin(c byte) (err error) {
	var (
		logp = `parseDirectiveOrigin`

		tok []byte
	)

	if c == ';' || c == '\n' || c == 0 {
		return fmt.Errorf(`%s: line %d: empty $origin directive`, logp, m.lineno)
	}

	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 {
		return fmt.Errorf(`%s: line %d: empty $origin directive`, logp, m.lineno)
	}

	m.zone.Origin = strings.ToLower(toDomainAbsolute(string(tok)))

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// parseDirectiveInclude parse the $INCLUDE directive in the following format,
//
//	$INCLUDE <file-name> [<domain-name>] [<comment>]
func (m *zoneParser) parseDirectiveInclude(c byte) (err error) {
	var (
		logp = `parseDirectiveInclude`

		tok []byte
	)

	if c == ';' || c == '\n' || c == 0 {
		return fmt.Errorf(`%s: line %d: empty $include directive`, logp, m.lineno)
	}

	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 {
		return fmt.Errorf(`%s: line %d: empty $include directive`, logp, m.lineno)
	}

	tok = ascii.ToLower(tok)

	var (
		incfile = string(tok)
		dname   = m.zone.Origin
	)

	// Check if include followed by domain name.
	if c == ' ' || c == '\t' {
		tok, c = m.parser.ReadNoSpace()
		if len(tok) != 0 {
			tok = ascii.ToLower(tok)
			dname = string(tok)
		}
	}

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var zoneFile *Zone

	zoneFile, err = ParseZoneFile(incfile, dname, m.zone.SOA.Minimum)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	m.zone.messages = append(m.zone.messages, zoneFile.messages...)

	return nil
}

// parseDirectiveTTL parse the $TTL directive in the following format,
//
//	$TTL <digits> [comment]
func (m *zoneParser) parseDirectiveTTL(c byte) (err error) {
	var (
		logp = `parseDirectiveTTL`

		stok string
		tok  []byte
	)

	if c == ';' || c == '\n' || c == 0 {
		return fmt.Errorf(`%s: line %d: empty $TTL directive`, logp, m.lineno)
	}

	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 {
		return fmt.Errorf(`%s: line %d: empty $TTL directive`, logp, m.lineno)
	}

	tok = ascii.ToLower(tok)
	stok = string(tok)

	m.zone.SOA.Minimum, err = parseTTL(tok, stok)
	if err != nil {
		return fmt.Errorf(`%s: line %d: %w`, logp, m.lineno, err)
	}

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// parseTTL convert characters of duration, with or without unit, to seconds.
func parseTTL(tok []byte, stok string) (seconds uint32, err error) {
	var (
		v   uint64
		dur time.Duration
	)

	if ascii.IsDigits(tok) {
		v, err = strconv.ParseUint(stok, 10, 32)
		if err != nil {
			return 0, fmt.Errorf(`invalid TTL value '%s'`, tok)
		}

		seconds = uint32(v)

		return seconds, nil
	}

	dur, err = libtime.ParseDuration(stok)
	if err != nil {
		return 0, fmt.Errorf(`invalid TTL value '%s'`, tok)
	}

	seconds = uint32(dur.Seconds())

	return seconds, nil
}

// parserRR parse the Resource Record (RR).
// If the RR start with blank, the prevRR is the previous line of RR and tok
// may contains TTL or class.
// If the RR does not start with blank, the prevRR is nil and tok contains the
// domain name.
//
// The RR contents take one of the following forms:
//
//	[<TTL>] [<class>] <type> <RDATA>
//
//	[<class>] [<TTL>] <type> <RDATA>
//
// The RR begins with optional TTL and class fields, followed by a type and
// RDATA field appropriate to the type and class.
// Class and type use the standard mnemonics, TTL is a decimal integer.
// Omitted class and TTL values are default to the last explicitly stated
// values.
func (m *zoneParser) parseRR(prevRR *ResourceRecord, tok []byte) (rr *ResourceRecord, err error) {
	var (
		logp = `parseRR`
		flag = flagRRStart

		stok   string
		ttl    uint32
		orgtok []byte
		c      byte
		ok     bool
	)

	rr = &ResourceRecord{
		TTL: m.zone.SOA.Minimum,
	}

	if prevRR == nil {
		rr.Name = m.generateDomainName(tok)
		if m.lastRR != nil {
			rr.Class = m.lastRR.Class
		} else {
			rr.Class = RecordClassIN
		}
	} else {
		rr.Name = prevRR.Name
		if prevRR.Type != RecordTypeSOA {
			rr.TTL = prevRR.TTL
		}
		rr.Class = prevRR.Class

		tok = ascii.ToUpper(tok)
		stok = string(tok)

		if ascii.IsDigit(tok[0]) {
			ttl, err = parseTTL(tok, stok)
			if err != nil {
				return nil, fmt.Errorf(`%s: line %d: invalid TTL '%s': %w`, logp, m.lineno, tok, err)
			}
			rr.TTL = ttl
			flag |= flagRRTtl
		} else {
			flag, ok = m.parseRRClassOrType(rr, stok, flag)
			if !ok {
				return nil, fmt.Errorf(`%s: line %d: unknown class or type '%s'`, logp, m.lineno, tok)
			}
		}
	}

	for flag != flagRREnd {
		tok, c = m.parser.ReadNoSpace()
		if len(tok) == 0 {
			return nil, fmt.Errorf(`%s: line %d: invalid RR statement '%s'`, logp, m.lineno, tok)
		}

		orgtok = libbytes.Copy(tok)
		tok = ascii.ToUpper(tok)
		stok = string(tok)

		switch flag {
		case flagRRStart:
			if ascii.IsDigit(tok[0]) {
				rr.TTL, err = parseTTL(tok, stok)
				if err != nil {
					return nil, fmt.Errorf(`%s: invalid TTL '%s': %w`, logp, tok, err)
				}
				flag |= flagRRTtl
				continue
			}

			fallthrough // If its not TTL, maybe class or type.

		case flagRRTtl:
			flag, ok = m.parseRRClassOrType(rr, stok, flag)
			if !ok {
				return nil, fmt.Errorf(`%s: line %d: unknown class or type '%s'`, logp, m.lineno, stok)
			}

		case flagRRClass:
			if ascii.IsDigit(tok[0]) {
				rr.TTL, err = parseTTL(tok, stok)
				if err != nil {
					return nil, err
				}
				flag |= flagRRTtl
				continue
			}

			fallthrough // If its not digit maybe type.

		case flagRRTtl | flagRRClass:
			rr.Type, ok = RecordTypes[stok]
			if !ok {
				return nil, fmt.Errorf(`%s: line %d: unknown class or type '%s'`, logp, m.lineno, stok)
			}
			flag = flagRRType

		case flagRRType:
			err = m.parseRRData(rr, orgtok, c)
			if err != nil {
				return nil, fmt.Errorf(`%s: line %d: %s`, logp, m.lineno, err)
			}
			flag = flagRREnd
		}
	}
	return rr, nil
}

// parseRRClassOrType check if token either class or type.
// It will return true if one of them is set, otherwise it will return false.
func (m *zoneParser) parseRRClassOrType(rr *ResourceRecord, stok string, flag int) (int, bool) {
	var ok bool

	rr.Class, ok = RecordClasses[stok]
	if ok {
		flag |= flagRRClass
		return flag, ok
	}

	// Set back to default class.
	rr.Class = RecordClassIN

	rr.Type, ok = RecordTypes[stok]
	if ok {
		flag = flagRRType
		return flag, ok
	}

	return flag, false
}

func (m *zoneParser) parseRRData(rr *ResourceRecord, tok []byte, c byte) (err error) {
	switch rr.Type {
	case RecordTypeA, RecordTypeAAAA:
		rr.Value = string(tok)
		err = m.skipLine(c)

	case RecordTypeNS, RecordTypeCNAME, RecordTypeMB, RecordTypeMG, RecordTypeMR, RecordTypePTR:
		rr.Value = m.generateDomainName(tok)
		err = m.skipLine(c)

	case RecordTypeSOA:
		err = m.parseSOA(rr, tok)

	// NULL RRs are not allowed in zone files.
	case RecordTypeNULL:
		err = fmt.Errorf("line %d: NULL type is not allowed", m.lineno)

	// In zone files, both ports and protocols are expressed using
	// mnemonics or decimal numbers.
	case RecordTypeWKS:
		// TODO(ms)
		_ = m.skipLine(c)

	case RecordTypeHINFO:
		err = m.parseHInfo(rr, tok)

	case RecordTypeMINFO:
		err = m.parseMInfo(rr, tok)

	case RecordTypeMX:
		err = m.parseMX(rr, tok)

	case RecordTypeTXT:
		err = m.parseTXT(rr, tok, c)

	case RecordTypeSRV:
		err = m.parseSRV(rr, tok)
	}

	return err
}

func (m *zoneParser) parseSOA(rr *ResourceRecord, tok []byte) (err error) {
	var (
		logp  = `parseSOA`
		rrSOA = &RDataSOA{}

		vint64      int64
		c           byte
		isMultiline bool
	)

	rr.Value = rrSOA

	tok = ascii.ToLower(tok)
	rrSOA.MName = m.generateDomainName(tok)

	// Get RNAME
	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 || c == 0 || c == ';' {
		return fmt.Errorf(`%s: line %d: incomplete SOA values`, logp, m.lineno)
	}

	tok = ascii.ToLower(tok)
	rrSOA.RName = m.generateDomainName(tok)

	// Get serial value.

	m.parser.AddDelimiters([]byte{'('})

	tok, c = m.parser.ReadNoSpace()
	if c == '(' {
		isMultiline = true
		m.parser.AddDelimiters([]byte{')'})
		if len(tok) == 0 {
			tok, c = m.getMultilineToken(m.parser)
		}
	}

	vint64, err = strconv.ParseInt(string(tok), 10, 64)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	rrSOA.Serial = uint32(vint64)

	// List of flag to know which part has been parsed.
	const (
		flagSerial  = 1
		flagRefresh = 2
		flagRetry   = 4
		flagExpire  = 8
		flagMinimum = 16
	)
	var flag = flagSerial

	for flag != flagMinimum {
		if isMultiline {
			tok, c = m.getMultilineToken(m.parser)
			if flag < flagExpire && c == ')' {
				return fmt.Errorf(`%s: line %d: incomplete SOA statement '%s'`, logp, m.lineno, tok)
			}
		} else {
			tok, c = m.parser.ReadNoSpace()
			if c == ';' {
				return fmt.Errorf(`%s: line %d: incomplete SOA statement '%s'`, logp, m.lineno, tok)
			}
		}
		if len(tok) == 0 || c == 0 {
			return fmt.Errorf(`%s: line %d: incomplete SOA statement '%s'`, logp, m.lineno, tok)
		}

		vint64, err = strconv.ParseInt(string(tok), 10, 64)
		if err != nil {
			return fmt.Errorf(`%s: line %d: invalid SOA value %s: %w`, logp, m.lineno, tok, err)
		}

		switch flag {
		case flagSerial:
			rrSOA.Refresh = int32(vint64)
			flag = flagRefresh

		case flagRefresh:
			rrSOA.Retry = int32(vint64)
			flag = flagRetry

		case flagRetry:
			rrSOA.Expire = int32(vint64)
			flag = flagExpire

		case flagExpire:
			rrSOA.Minimum = uint32(vint64)
			flag = flagMinimum
		}
	}

	if isMultiline {
		for c != ')' {
			tok, c = m.getMultilineToken(m.parser)
			if len(tok) != 0 {
				return fmt.Errorf(`%s: line %d: unknown token '%s'`, logp, m.lineno, tok)
			}
			if c == 0 {
				break
			}
		}
		if c != ')' {
			return fmt.Errorf(`%s: line %d: missing SOA closing parentheses`, logp, m.lineno)
		}
		m.parser.RemoveDelimiters([]byte{'(', ')'})
	}

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// getMultilineToken get the next token skipping empty or commented line.
func (m *zoneParser) getMultilineToken(parser *libbytes.Parser) (tok []byte, c byte) {
	for {
		tok, c = parser.ReadNoSpace()
		if c == ';' || c == '\n' {
			c = parser.SkipLine()
			m.lineno++
		}
		if len(tok) != 0 || c == 0 || c == ')' {
			break
		}
	}
	return tok, c
}

func (m *zoneParser) parseHInfo(rr *ResourceRecord, tok []byte) (err error) {
	var (
		logp    = `parseHInfo`
		rrHInfo = &RDataHINFO{}

		c byte
	)

	rr.Value = rrHInfo

	rrHInfo.CPU = tok

	// Get OS
	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 {
		return fmt.Errorf(`%s: line %d: missing HInfo OS value`, logp, m.lineno)
	}

	rrHInfo.OS = tok

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

func (m *zoneParser) parseMInfo(rr *ResourceRecord, tok []byte) (err error) {
	var (
		logp    = `parseMInfo`
		rrMInfo = &RDataMINFO{}

		c byte
	)

	rr.Value = rrMInfo

	rrMInfo.RMailBox = m.generateDomainName(tok)

	// Get EmailBox value
	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 {
		return fmt.Errorf(`%s: line %d: missing MInfo EmailBox value`, logp, m.lineno)
	}

	rrMInfo.EmailBox = m.generateDomainName(tok)

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

func (m *zoneParser) parseMX(rr *ResourceRecord, tok []byte) (err error) {
	var (
		logp = `parseMX`

		rrMX *RDataMX
		pref int64
		c    byte
	)

	pref, err = strconv.ParseInt(string(tok), 10, 64)
	if err != nil {
		return fmt.Errorf(`%s: line %d: invalid MX Preference '%s': %w`, logp, m.lineno, tok, err)
	}

	rrMX = &RDataMX{
		Preference: int16(pref),
	}
	rr.Value = rrMX

	// Get Exchange value
	tok, c = m.parser.ReadNoSpace()
	if len(tok) == 0 {
		return fmt.Errorf(`%s: line %d: missing MX Exchange value`, logp, m.lineno)
	}

	rrMX.Exchange = m.generateDomainName(tok)

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// parseTXT parse TXT resource data (rdata).
// For rdata that contains space, it must be double quote or use the encoding
// \DDD. for example, given the following rdata "a page", it could be encoded
// as
//
//	TXT "a page"
//	TXT a\032page
func (m *zoneParser) parseTXT(rr *ResourceRecord, txt []byte, c byte) (err error) {
	var (
		logp        = `parseTXT`
		spaceDelims = []byte{' ', '\t'}
		quoteDelims = []byte{'"'}

		tok []byte
	)

	if txt[0] == '"' {
		if c == ' ' || c == '\t' {
			txt = append(txt, c)
		}

		m.parser.RemoveDelimiters(spaceDelims)
		m.parser.AddDelimiters(quoteDelims)

		tok, c = m.parser.Read()
		if c != '"' {
			return fmt.Errorf(`%s: missing closing '"'`, logp)
		}
		txt = append(txt, tok...)
		txt = append(txt, '"')

		m.parser.RemoveDelimiters(quoteDelims)
		m.parser.AddDelimiters(spaceDelims)
	}

	m.parser.SkipLine()
	m.lineno++

	txt, err = m.decodeString(txt)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	rr.Value = string(txt)

	return nil
}

// parseSRV parse the SRV record.
// The tok parameter contains the unparsed field for priority.
func (m *zoneParser) parseSRV(rr *ResourceRecord, tok []byte) (err error) {
	var (
		logp         = `parseSRV`
		svcProtoName = strings.SplitN(rr.Name, `.`, 3)
	)
	if len(svcProtoName) <= 1 {
		return fmt.Errorf(`%s: line %d: invalid _service._proto.name: %s`, logp, m.lineno, rr.Name)
	}

	var (
		rrSRV = &RDataSRV{
			Service: svcProtoName[0],
			Proto:   svcProtoName[1],
		}
	)

	if len(svcProtoName) == 2 {
		rrSRV.Name = m.zone.Origin
	} else {
		rrSRV.Name = svcProtoName[2]
	}

	const (
		flagPriority = iota
		flagWeight
		flagPort
		flagTarget
	)
	var (
		flag int
		stok string
		vint int
		c    byte
	)

	stok = string(tok)
	vint, err = strconv.Atoi(stok)
	if err != nil {
		return fmt.Errorf(`%s: line %d: invalid Priority value %s: %w`, logp, m.lineno, tok, err)
	}
	rrSRV.Priority = uint16(vint)
	flag = flagPriority

	for flag != flagTarget {
		tok, c = m.parser.ReadNoSpace()
		if len(tok) == 0 {
			return fmt.Errorf(`%s: line %d: incomplete SRV RDATA`, logp, m.lineno)
		}

		stok = string(tok)

		switch flag {
		case flagPriority:
			vint, err = strconv.Atoi(stok)
			if err != nil {
				return fmt.Errorf(`%s: line %d: invalid Weight value %s: %w`, logp, m.lineno, tok, err)
			}
			rrSRV.Weight = uint16(vint)
			flag = flagWeight

		case flagWeight:
			vint, err = strconv.Atoi(stok)
			if err != nil {
				return fmt.Errorf(`%s: line %d: invalid Port value %s: %w`, logp, m.lineno, tok, err)
			}
			rrSRV.Port = uint16(vint)
			flag = flagPort

		case flagPort:
			rrSRV.Target = string(tok)
			flag = flagTarget
		}
	}

	err = m.skipLine(c)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	rr.Value = rrSRV

	return nil
}

func (m *zoneParser) skipLine(c byte) (err error) {
	var tok []byte

	if c == ' ' || c == '\t' {
		tok, c = m.parser.ReadNoSpace()
		if len(tok) != 0 {
			return fmt.Errorf(`unknown token '%s'`, tok)
		}
	}
	switch c {
	case 0, '\n':
	case ';':
		m.parser.SkipLine()
	}
	m.lineno++

	return nil
}

// decodeString decode a [character-string].
//
// A character-string is expressed in one or two ways: as a contiguous set of
// characters without interior spaces, or as a string beginning with a " and
// ending with a ".
//
// For contiguous string without double quotes, the following encoding are
// recognized,
//
//   - \X where X is any character other than a digit (0-9), is used to quote
//     that character so that its special meaning does not apply.
//     For example, "\." can be used to place a dot character in a label.
//   - \DDD  where each D is a digit is the octet corresponding to the decimal
//     number described by DDD.
//     The resulting octet is assumed to be text and is not checked for
//     special meaning.
//
// For quoted string, inside a " delimited string any character can occur,
// except for a " itself, which must be quoted using \ (back slash).
//
// [character-string]: https://datatracker.ietf.org/doc/html/rfc1035#section-5.1
func (m *zoneParser) decodeString(in []byte) (out []byte, err error) {
	var (
		logp = `decodeString`
		size = len(in)

		c     byte
		isEsc bool
	)

	out = make([]byte, 0, len(in))
	if in[0] == '"' && in[size-1] == '"' {
		// Un-escape the backslash quote only
		for _, c = range in[1 : size-1] {
			if isEsc {
				if c == '"' {
					out = append(out, '"')
				} else {
					out = append(out, '\\', c)
				}
				isEsc = false
				continue
			}
			if c == '\\' {
				isEsc = true
				continue
			}
			out = append(out, c)
		}
		return out, nil
	}

	var x int
	for x = 0; x < size; x++ {
		c = in[x]
		if ascii.IsSpace(c) {
			break
		}
		if isEsc {
			if !ascii.IsDigit(c) {
				out = append(out, c)
				isEsc = false
				continue
			}

			var digits = make([]byte, 0, 3)

			for x < size && len(digits) <= 2 {
				digits = append(digits, in[x])
				if !ascii.IsDigit(in[x]) {
					return nil, fmt.Errorf(`%s: invalid digits: \%s`, logp, digits)
				}
				x++
			}
			if len(digits) != 3 {
				return nil, fmt.Errorf(`%s: invalid digits length: \%s`, logp, digits)
			}
			x--

			var vint int64
			vint, err = strconv.ParseInt(string(digits), 10, 8)
			if err != nil {
				return nil, fmt.Errorf(`%s: invalid octet: \%s`, logp, digits)
			}

			out = append(out, byte(vint))
			isEsc = false
			continue
		}
		if c == '\\' {
			isEsc = true
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (m *zoneParser) generateDomainName(dname []byte) (out string) {
	dname = ascii.ToLower(dname)
	if bytes.Equal(dname, []byte("@")) {
		return m.zone.Origin
	}
	if dname[len(dname)-1] == '.' {
		return string(dname)
	}
	out = string(dname) + "." + m.zone.Origin
	return out
}

// push resource record (RR) into message answer only if domain name, type,
// and class already exist; otherwise it will create new message with question
// based on RR.
func (m *zoneParser) push(rr *ResourceRecord) error {
	m.lastRR = rr
	return m.zone.add(rr)
}

func (m *zoneParser) setMinimumTTL() {
	var (
		msg *Message
		x   int
	)

	for _, msg = range m.zone.messages {
		for x = 0; x < len(msg.Answer); x++ {
			if msg.Answer[x].TTL < m.zone.SOA.Minimum {
				msg.Answer[x].TTL = m.zone.SOA.Minimum
			}
		}
		for x = 0; x < len(msg.Authority); x++ {
			if msg.Authority[x].TTL < m.zone.SOA.Minimum {
				msg.Authority[x].TTL = m.zone.SOA.Minimum
			}
		}
		for x = 0; x < len(msg.Additional); x++ {
			if msg.Additional[x].TTL < m.zone.SOA.Minimum {
				msg.Additional[x].TTL = m.zone.SOA.Minimum
			}
		}
	}
}

func (m *zoneParser) pack() {
	var (
		msg *Message
		err error
	)

	for _, msg = range m.zone.messages {
		msg.Header.ANCount = uint16(len(msg.Answer))
		msg.Header.NSCount = uint16(len(msg.Authority))
		msg.Header.ARCount = uint16(len(msg.Additional))

		_, err = msg.Pack()
		if err != nil {
			msg.Header.ANCount = 0
		}
	}
}
