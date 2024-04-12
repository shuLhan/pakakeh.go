// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
	"git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
)

// Message represent a DNS message.
//
// All communications inside of the domain protocol are carried in a single
// format called a message.  The top level format of message is divided
// into 5 sections (some of which are empty in certain cases) shown below:
//
//	+---------------------+
//	|        Header       |
//	+---------------------+
//	|       Question      | the question for the name server
//	+---------------------+
//	|        Answer       | RRs answering the question
//	+---------------------+
//	|      Authority      | RRs pointing toward an authority
//	+---------------------+
//	|      Additional     | RRs holding additional information
//	+---------------------+
//
// The names of the sections after the header are derived from their use in
// standard queries.  The question section contains fields that describe a
// question to a name server.  These fields are a query type (QTYPE), a
// query class (QCLASS), and a query domain name (QNAME).  The last three
// sections have the same format: a possibly empty list of concatenated
// resource records (RRs).  The answer section contains RRs that answer the
// question; the authority section contains RRs that point toward an
// authoritative name server; the additional records section contains RRs
// which relate to the query, but are not strictly answers for the
// question. [1]
//
// [1] RFC 1035 - 4.1. Format
type Message struct {
	dnameOff map[string]uint16
	dname    string

	Answer     []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
	packet     []byte

	Question MessageQuestion
	Header   MessageHeader
}

// NewMessage create, initialize, and return new message.
func NewMessage() *Message {
	return &Message{
		Header: MessageHeader{
			IsQuery: true,
			IsRD:    true,
			QDCount: 1,
		},
		Question: MessageQuestion{
			Type:  RecordTypeA,
			Class: RecordClassIN,
		},
		dnameOff: make(map[string]uint16),
	}
}

// NewMessageAddress create new DNS message for hostname that contains one or
// more A or AAAA addresses.
// The addresses must be all IPv4 or IPv6, the first address define the query
// type.
// If hname is not valid hostname or one of the address is not valid IP
// address it will return nil.
func NewMessageAddress(hname []byte, addresses [][]byte) (msg *Message) {
	if !libnet.IsHostnameValid(hname, false) {
		return nil
	}
	if len(addresses) == 0 {
		return nil
	}

	var (
		addr  = addresses[0]
		rtype = RecordTypeFromAddress(addr)

		rr  ResourceRecord
		err error
	)
	if rtype == 0 {
		return nil
	}

	hname = ascii.ToLower(hname)

	rr = ResourceRecord{
		Name:  string(hname),
		Type:  rtype,
		Class: RecordClassIN,
		TTL:   defaultTTL,
		Value: string(addr),
	}

	msg = &Message{
		Header: MessageHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: MessageQuestion{
			Name:  string(hname),
			Type:  rtype,
			Class: RecordClassIN,
		},
		Answer: []ResourceRecord{rr},
	}

	for _, addr = range addresses[1:] {
		rtype = RecordTypeFromAddress(addr)
		if rtype == 0 {
			continue
		}
		if rtype != msg.Question.Type {
			continue
		}
		msg.Answer = append(msg.Answer, ResourceRecord{
			Name:  string(hname),
			Type:  rtype,
			Class: RecordClassIN,
			TTL:   defaultTTL,
			Value: string(addr),
		})
		msg.Header.ANCount++
	}

	_, err = msg.Pack()
	if err != nil {
		return nil
	}

	return msg
}

// NewMessageFromRR create new message with one RR as an answer.
func NewMessageFromRR(rr *ResourceRecord) (msg *Message, err error) {
	msg = &Message{
		Header: MessageHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: MessageQuestion{
			Name:  rr.Name,
			Type:  rr.Type,
			Class: rr.Class,
		},
		Answer: []ResourceRecord{*rr},
	}
	_, err = msg.Pack()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Unpack the raw DNS packet into a Message.
func UnpackMessage(packet []byte) (msg *Message, err error) {
	var logp = `UnpackMessage`

	msg = &Message{
		packet: packet,
	}

	err = msg.UnpackHeaderQuestion()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w: %w`, logp, errInvalidMessage, err)
	}

	var (
		startIdx = uint(sectionHeaderSize + msg.Question.size())

		rr ResourceRecord
		x  uint16
	)
	for ; x < msg.Header.ANCount; x++ {
		rr = ResourceRecord{}

		startIdx, err = rr.unpack(msg.packet, startIdx)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w: %w`, logp, errInvalidMessage, err)
		}

		msg.Answer = append(msg.Answer, rr)
	}

	for x = 0; x < msg.Header.NSCount; x++ {
		rr = ResourceRecord{}

		startIdx, err = rr.unpack(msg.packet, startIdx)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w: %w`, logp, errInvalidMessage, err)
		}
		msg.Authority = append(msg.Authority, rr)
	}

	for x = 0; x < msg.Header.ARCount; x++ {
		rr = ResourceRecord{}

		startIdx, err = rr.unpack(msg.packet, startIdx)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w: %w`, logp, errInvalidMessage, err)
		}

		msg.Additional = append(msg.Additional, rr)
	}

	return msg, nil
}

// AddAnswer to the Answer field and re-pack it again.
func (msg *Message) AddAnswer(rr *ResourceRecord) (err error) {
	switch rr.Type {
	case RecordTypeSOA, RecordTypePTR:
		if len(msg.Answer) > 0 {
			msg.Answer[0] = *rr
		} else {
			msg.Answer = append(msg.Answer, *rr)
		}
	default:
		msg.Answer = append(msg.Answer, *rr)
		msg.Header.ANCount++
	}

	_, err = msg.Pack()

	return err
}

// AddAuthority add the rr to list of Authority.
// Calling this method mark the message as answer, instead of query.
//
// If the rr is SOA, it will replace the existing record if exist and set
// the flag authoritative answer (IsAA) in header to true.
// If the rr is NS, it will be added only if its not exist.
//
// It will return an error if the rr type is not SOA or NS or the size of
// records in Authority is full, maximum four records.
func (msg *Message) AddAuthority(rr *ResourceRecord) (err error) {
	var (
		logp = `AddAuthority`

		rrAuth ResourceRecord
		x      int
		found  bool
	)

	switch rr.Type {
	case RecordTypeSOA:
		for x, rrAuth = range msg.Authority {
			if rrAuth.Type == RecordTypeSOA {
				msg.Authority[x] = *rr
				found = true
				break
			}
		}
	case RecordTypeNS:
		var (
			rrVal string
			val   string
			ok    bool
		)

		rrVal, ok = rr.Value.(string)
		if !ok {
			return fmt.Errorf(`%s: expecting NS value as string, got %T`, logp, rr.Value)
		}

		for _, rrAuth = range msg.Authority {
			if rrAuth.Type != RecordTypeNS {
				continue
			}
			val, ok = rrAuth.Value.(string)
			if !ok {
				return fmt.Errorf(`%s: expecting NS value as string, got %T`, logp, rr.Value)
			}
			if rrVal == val {
				found = true
				break
			}
		}

	default:
		return fmt.Errorf(`%s: expecting SOA or NS, got %s`, logp, RecordTypeNames[rr.Type])
	}

	if !found {
		if len(msg.Authority) >= 4 {
			return fmt.Errorf(`%s: records is full`, logp)
		}
		msg.Authority = append(msg.Authority, *rr)
	}

	msg.Header.NSCount = uint16(len(msg.Authority))
	msg.SetQuery(false)

	if rr.Type == RecordTypeSOA {
		msg.SetAuthorativeAnswer(true)
	}

	_, err = msg.Pack()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// FilterAnswers return resource record in Answer that match only with
// specific query type.
func (msg *Message) FilterAnswers(t RecordType) (answers []ResourceRecord) {
	var rr ResourceRecord
	for _, rr = range msg.Answer {
		if rr.Type == t {
			answers = append(answers, rr)
		}
	}
	return
}

func (msg *Message) compress() bool {
	if len(msg.dname) == 0 || msg.dname == "." {
		return false
	}

	var (
		off uint16
		ok  bool
	)
	off, ok = msg.dnameOff[msg.dname]
	if ok {
		msg.packet = append(msg.packet, maskPointer|byte(off>>8))
		msg.packet = append(msg.packet, byte(off))
		return true
	}
	return false
}

// packDomainName convert string of domain-name into DNS domain-name format.
func (msg *Message) packDomainName(dname []byte, doCompress bool) (n int) {
	var ok bool

	dname = ascii.ToLower(dname)
	msg.dname = string(dname)

	if doCompress {
		ok = msg.compress()
		if ok {
			return 2
		}
	}

	var (
		count = byte(0)

		idxCount int
		c        byte
		x        int
	)

	msg.packet = append(msg.packet, 0)
	idxCount = len(msg.packet) - 1
	msg.dnameOff[msg.dname] = uint16(idxCount)
	n++

	for x = 0; x < len(dname); x++ {
		c = dname[x]

		if c == '\\' {
			x++
			if x == len(dname) {
				return n
			}

			c = dname[x]

			// \DDD  where each D is a digit is the octet
			// corresponding to the decimal number described by
			// DDD.  The resulting octet is assumed to be text and
			// is not checked for special meaning.
			if ascii.IsDigit(c) {
				if x+2 >= len(dname) {
					return n
				}
				var d int
				d, _ = strconv.Atoi(string(dname[x : x+3]))
				c = byte(d)
				if c >= 'A' && c <= 'Z' {
					c += 32
				}
				x += 2
			}
			msg.packet = append(msg.packet, c)
			count++
			continue
		}
		if c == '.' {
			// Skip name that prefixed with '.', e.g.
			// '...test.com'
			if count == 0 {
				continue
			}

			msg.packet[idxCount] = count

			msg.dname = string(dname[x+1:])
			n += int(count)

			if doCompress {
				ok = msg.compress()
				if ok {
					n += 2
					return n
				}
			}

			count = 0
			msg.packet = append(msg.packet, 0)
			idxCount = len(msg.packet) - 1
			n++

			if len(msg.dname) == 0 || msg.dname == `.` {
				dname = nil
				break
			}

			msg.dnameOff[msg.dname] = uint16(idxCount)
			continue
		}

		msg.packet = append(msg.packet, c)
		count++
	}
	if count > 0 {
		msg.packet[idxCount] = count
		n += int(count)

		msg.packet = append(msg.packet, 0)
		n++
	}

	return n
}

func (msg *Message) packQuestion() {
	msg.packDomainName([]byte(msg.Question.Name), false)
	msg.packet = libbytes.AppendUint16(msg.packet, uint16(msg.Question.Type))
	msg.packet = libbytes.AppendUint16(msg.packet, uint16(msg.Question.Class))
}

func (msg *Message) packRR(rr *ResourceRecord) {
	var (
		rrOPT *RDataOPT
	)

	if rr.Type == RecordTypeOPT {
		// MUST be 0 (root domain).
		msg.packet = append(msg.packet, 0)
		rrOPT, _ = rr.Value.(*RDataOPT)
	} else {
		msg.packDomainName([]byte(rr.Name), true)
	}

	msg.packet = libbytes.AppendUint16(msg.packet, uint16(rr.Type))
	msg.packet = libbytes.AppendUint16(msg.packet, uint16(rr.Class))

	if rr.Type == RecordTypeOPT {
		rr.TTL = 0

		// Pack extended code and version to TTL
		rr.TTL = uint32(rrOPT.ExtRCode) << 24
		rr.TTL |= (uint32(rrOPT.Version) << 16)

		if rrOPT.DO {
			rr.TTL |= maskOPTDO
		}
	}

	rr.idxTTL = uint16(len(msg.packet))
	msg.packet = libbytes.AppendUint32(msg.packet, rr.TTL)

	msg.packRData(rr)
}

func (msg *Message) packRData(rr *ResourceRecord) {
	switch rr.Type {
	case RecordTypeA:
		msg.packA(rr)
	case RecordTypeNS:
		msg.packTextAsDomain(rr)
	case RecordTypeMD:
		// obsolete
	case RecordTypeMF:
		// obsolete
	case RecordTypeCNAME:
		msg.packTextAsDomain(rr)
	case RecordTypeSOA:
		msg.packSOA(rr)
	case RecordTypeMB:
		msg.packTextAsDomain(rr)
	case RecordTypeMG:
		msg.packTextAsDomain(rr)
	case RecordTypeMR:
		msg.packTextAsDomain(rr)
	case RecordTypeNULL:
		msg.packTextAsDomain(rr)
	case RecordTypeWKS:
		msg.packWKS(rr)
	case RecordTypePTR:
		msg.packTextAsDomain(rr)
	case RecordTypeHINFO:
		msg.packHINFO(rr)
	case RecordTypeMINFO:
		msg.packMINFO(rr)
	case RecordTypeMX:
		msg.packMX(rr)
	case RecordTypeTXT:
		msg.packTXT(rr)
	case RecordTypeSRV:
		msg.packSRV(rr)
	case RecordTypeAAAA:
		msg.packAAAA(rr)
	case RecordTypeOPT:
		msg.packOPT(rr)
	case RecordTypeSVCB:
		msg.packSVCB(rr)
	case RecordTypeHTTPS:
		msg.packHTTPS(rr)
	}
}

func (msg *Message) packA(rr *ResourceRecord) {
	msg.packet = libbytes.AppendUint16(msg.packet, rdataIPv4Size)

	var (
		rrText string
		ip     net.IP
		ipv4   net.IP
	)

	rrText, _ = rr.Value.(string)

	ip = net.ParseIP(rrText)
	if ip == nil {
		msg.packet = append(msg.packet, rrText[:rdataIPv4Size]...)
	} else {
		ipv4 = ip.To4()
		if ipv4 == nil {
			msg.packet = append(msg.packet, ip[:rdataIPv4Size]...)
		} else {
			msg.packet = append(msg.packet, ipv4...)
		}
	}
}

func (msg *Message) packTextAsDomain(rr *ResourceRecord) {
	var (
		off       = uint(len(msg.packet))
		rrText, _ = rr.Value.(string)

		n int
	)

	// Reserve two octets for rdlength
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	n = msg.packDomainName([]byte(rrText), true)
	libbytes.WriteUint16(msg.packet, off, uint16(n))
}

func (msg *Message) packSOA(rr *ResourceRecord) {
	var (
		off      = uint(len(msg.packet))
		rrSOA, _ = rr.Value.(*RDataSOA)

		n     int
		total int
	)

	// Reserve two octets for rdlength.
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	n = msg.packDomainName([]byte(rrSOA.MName), true)
	total = n
	n = msg.packDomainName([]byte(rrSOA.RName), true)
	total += n

	msg.packet = libbytes.AppendUint32(msg.packet, rrSOA.Serial)
	msg.packet = libbytes.AppendInt32(msg.packet, rrSOA.Refresh)
	msg.packet = libbytes.AppendInt32(msg.packet, rrSOA.Retry)
	msg.packet = libbytes.AppendInt32(msg.packet, rrSOA.Expire)
	msg.packet = libbytes.AppendUint32(msg.packet, rrSOA.Minimum)
	total += 20

	// Write rdlength.
	libbytes.WriteUint16(msg.packet, off, uint16(total))
}

func (msg *Message) packWKS(rr *ResourceRecord) {
	var (
		rrWKS, _ = rr.Value.(*RDataWKS)
		n        = uint16(5 + len(rrWKS.BitMap))
	)

	// Write rdlength.
	msg.packet = libbytes.AppendUint16(msg.packet, n)

	msg.packet = append(msg.packet, rrWKS.Address[:4]...)
	msg.packet = append(msg.packet, rrWKS.Protocol)
	msg.packet = append(msg.packet, rrWKS.BitMap...)
}

func (msg *Message) packHINFO(rr *ResourceRecord) {
	var (
		rrHInfo, _ = rr.Value.(*RDataHINFO)
		n          = len(rrHInfo.CPU) + 1
	)

	// Write rdlength.
	n += len(rrHInfo.OS) + 1
	msg.packet = libbytes.AppendUint16(msg.packet, uint16(n))
	msg.packet = append(msg.packet, byte(len(rrHInfo.CPU)))
	msg.packet = append(msg.packet, rrHInfo.CPU...)
	msg.packet = append(msg.packet, byte(len(rrHInfo.OS)))
	msg.packet = append(msg.packet, rrHInfo.OS...)
}

func (msg *Message) packMINFO(rr *ResourceRecord) {
	var (
		rrMInfo, _ = rr.Value.(*RDataMINFO)
		off        = uint(len(msg.packet))

		n int
	)

	// Reserve two octets for rdlength.
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	n = msg.packDomainName([]byte(rrMInfo.RMailBox), true)
	n += msg.packDomainName([]byte(rrMInfo.EmailBox), true)

	// Write rdlength.
	libbytes.WriteUint16(msg.packet, off, uint16(n))
}

func (msg *Message) packMX(rr *ResourceRecord) {
	var (
		rrMX, _ = rr.Value.(*RDataMX)

		off uint
		n   int
	)

	// Reserve two octets for rdlength.
	off = uint(len(msg.packet))
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	msg.packet = libbytes.AppendInt16(msg.packet, rrMX.Preference)

	n = msg.packDomainName([]byte(rrMX.Exchange), true)

	// Write rdlength.
	libbytes.WriteUint16(msg.packet, off, uint16(n+2))
}

func (msg *Message) packTXT(rr *ResourceRecord) {
	var (
		rrText, _ = rr.Value.(string)
		n         = uint16(len(rrText))
	)

	msg.packet = libbytes.AppendUint16(msg.packet, n+1)

	msg.packet = append(msg.packet, byte(n))
	msg.packet = append(msg.packet, rrText...)
}

func (msg *Message) packSRV(rr *ResourceRecord) {
	var (
		rrSRV, _ = rr.Value.(*RDataSRV)
		off      = uint(len(msg.packet))

		n int
	)

	// Reserve two octets for rdlength
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	msg.packet = libbytes.AppendUint16(msg.packet, rrSRV.Priority)
	msg.packet = libbytes.AppendUint16(msg.packet, rrSRV.Weight)
	msg.packet = libbytes.AppendUint16(msg.packet, rrSRV.Port)

	n = msg.packDomainName([]byte(rrSRV.Target), false) + 6

	// Write rdlength.
	libbytes.WriteUint16(msg.packet, off, uint16(n))
}

func (msg *Message) packAAAA(rr *ResourceRecord) {
	var (
		rrText, _ = rr.Value.(string)
		ip        = net.ParseIP(rrText)
	)

	msg.packet = libbytes.AppendUint16(msg.packet, rdataIPv6Size)

	if ip == nil {
		msg.packet = append(msg.packet, rrText[:rdataIPv6Size]...)
	} else {
		msg.packet = append(msg.packet, ip...)
	}
}

// packOPT pack the RDLEN and RDATA of OPT record.
func (msg *Message) packOPT(rr *ResourceRecord) {
	var rrOPT *RDataOPT

	rrOPT, _ = rr.Value.(*RDataOPT)

	var rdata = rrOPT.pack()
	msg.packet = libbytes.AppendUint16(msg.packet, uint16(len(rdata)))
	msg.packet = append(msg.packet, rdata...)
}

func (msg *Message) packSVCB(rr *ResourceRecord) {
	var (
		svcb *RDataSVCB
		ok   bool
	)

	svcb, ok = rr.Value.(*RDataSVCB)
	if !ok {
		return
	}

	// Reserve two octets for rdlength.
	var off = uint(len(msg.packet))
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	var n = svcb.pack(msg)

	// Write rdlength.
	libbytes.WriteUint16(msg.packet, off, uint16(n))
}

func (msg *Message) packHTTPS(rr *ResourceRecord) {
	var (
		rrhttps *RDataHTTPS
		ok      bool
	)

	rrhttps, ok = rr.Value.(*RDataHTTPS)
	if !ok {
		return
	}

	// Reserve two octets for rdlength.
	var off = uint(len(msg.packet))
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	// Priority.
	msg.packet = libbytes.AppendUint16(msg.packet, 0)

	var n = msg.packDomainName([]byte(rrhttps.TargetName), false)

	// In HTTPS (AliasMode), Params is ignored.

	// Write rdlength.
	libbytes.WriteUint16(msg.packet, off, uint16(n))
}

func (msg *Message) packIPv4(addr string) {
	var ip = net.ParseIP(addr)
	if ip == nil {
		msg.packet = append(msg.packet, []byte{0, 0, 0, 0}...)
	} else {
		var ipv4 = ip.To4()
		if ipv4 == nil {
			msg.packet = append(msg.packet, []byte{0, 0, 0, 0}...)
		} else {
			msg.packet = append(msg.packet, ipv4...)
		}
	}
}

func (msg *Message) packIPv6(addr string) {
	var ip = net.ParseIP(addr)

	if ip == nil {
		msg.packet = append(msg.packet, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
	} else {
		msg.packet = append(msg.packet, ip...)
	}
}

// Reset the message fields.
func (msg *Message) Reset() {
	msg.Header.Reset()
	msg.Question.Reset()

	msg.ResetRR()
	msg.packet = nil

	msg.dname = ""
	msg.dnameOff = make(map[string]uint16)
}

// ResetRR free allocated resource records in message.  This function can be
// used to release some memory after message has been packed, but the raw
// packet may still be in use.
func (msg *Message) ResetRR() {
	msg.Answer = nil
	msg.Authority = nil
	msg.Additional = nil
}

// IsExpired will return true if at least one resource record in answers is
// expired, where their TTL value is equal to 0.
// As long as the answers RR exist and no TTL is 0, it will return false.
//
// If RR answers is empty, then the TTL on authority RR will be checked for
// zero.
//
// There is no check to be done on additional RR, since its may contain EDNS
// with zero TTL.
func (msg *Message) IsExpired() bool {
	var (
		x int
	)

	for x = 0; x < len(msg.Answer); x++ {
		if msg.Answer[x].TTL == 0 {
			return true
		}
	}
	if len(msg.Answer) > 0 {
		return false
	}

	for x = 0; x < len(msg.Authority); x++ {
		if msg.Authority[x].TTL == 0 {
			return true
		}
	}

	return false
}

// Pack convert message into datagram packet.  The result of packing
// a message will be saved in Packet field and returned.
func (msg *Message) Pack() ([]byte, error) {
	msg.dnameOff = make(map[string]uint16)
	msg.packet = msg.packet[:0]

	msg.Header.ANCount = uint16(len(msg.Answer))
	msg.Header.NSCount = uint16(len(msg.Authority))
	msg.Header.ARCount = uint16(len(msg.Additional))

	var (
		header = msg.Header.pack()

		x int
	)

	msg.packet = append(msg.packet, header...)

	msg.packQuestion()

	if msg.Header.IsQuery {
		msg.dnameOff = nil
		return msg.packet, nil
	}

	for x = 0; x < len(msg.Answer); x++ {
		msg.packRR(&msg.Answer[x])
	}
	for x = 0; x < len(msg.Authority); x++ {
		msg.packRR(&msg.Authority[x])
	}
	for x = 0; x < len(msg.Additional); x++ {
		msg.packRR(&msg.Additional[x])
	}

	msg.dnameOff = nil

	return msg.packet, nil
}

// RemoveAnswer remove the RR from list of answer.
func (msg *Message) RemoveAnswer(rrIn *ResourceRecord) (*ResourceRecord, error) {
	var (
		rrAnswer ResourceRecord
		err      error
		x        int
	)

	for x, rrAnswer = range msg.Answer {
		if !reflect.IsEqual(rrAnswer.Value, rrIn.Value) {
			continue
		}
		copy(msg.Answer[x:], msg.Answer[x+1:])
		msg.Answer = msg.Answer[:len(msg.Answer)-1]
		msg.Header.ANCount--
		_, err = msg.Pack()
		if err != nil {
			return nil, err
		}
		return &rrAnswer, nil
	}
	return nil, nil
}

// SetAuthorativeAnswer set the header authoritative answer to true (1) or
// false (0).
func (msg *Message) SetAuthorativeAnswer(isAA bool) {
	msg.Header.IsAA = isAA
	if len(msg.packet) > 2 {
		if isAA {
			msg.packet[2] |= headerIsAA
		} else {
			msg.packet[2] = (msg.packet[2] & 0xFB)
		}
	}
}

// SetID in section header and in packet.
func (msg *Message) SetID(id uint16) {
	msg.Header.ID = id
	if len(msg.packet) > 2 {
		libbytes.WriteUint16(msg.packet, 0, id)
	}
}

// SetQuery set the message as query (0) or as response (1) in header and in
// packet.
// Setting the message as query will also turning off AA, TC, and RA flags.
func (msg *Message) SetQuery(isQuery bool) {
	msg.Header.IsQuery = isQuery
	if len(msg.packet) > 3 {
		if isQuery {
			// Turn off query, authoritative answer, and truncated
			// flags.
			msg.packet[2] &= 0x71
			// Turn off recursion available flag.
			msg.packet[3] &= 0x7F
		} else {
			msg.packet[2] |= headerIsResponse
		}
	}
}

// SetRecursionDesired set the message to allow recursion (true=1) or
// not (false=0) in header and packet.
func (msg *Message) SetRecursionDesired(isRD bool) {
	msg.Header.IsRD = isRD
	if len(msg.packet) > 2 {
		if isRD {
			msg.packet[2] |= headerIsRD
		} else {
			msg.packet[2] &= 0xFE
		}
	}
}

// SetResponseCode in message header and in packet.
func (msg *Message) SetResponseCode(code ResponseCode) {
	msg.Header.RCode = code
	if len(msg.packet) > 3 {
		if code == RCodeOK {
			msg.packet[3] &= 0xF0
		} else {
			msg.packet[3] |= (0x0F & byte(code))
		}
	}
}

// SubTTL subtract TTL in each resource records and in packet by n seconds.
// If TTL is less than n, it will set to 0.
func (msg *Message) SubTTL(n uint32) {
	var (
		x int
	)

	for x = 0; x < len(msg.Answer); x++ {
		if msg.Answer[x].TTL < n {
			msg.Answer[x].TTL = 0
		} else {
			msg.Answer[x].TTL -= n
		}
		libbytes.WriteUint32(msg.packet, uint(msg.Answer[x].idxTTL), msg.Answer[x].TTL)
	}
	for x = 0; x < len(msg.Authority); x++ {
		if msg.Authority[x].TTL < n {
			msg.Authority[x].TTL = 0
		} else {
			msg.Authority[x].TTL -= n
		}
		libbytes.WriteUint32(msg.packet, uint(msg.Authority[x].idxTTL), msg.Authority[x].TTL)
	}
	for x = 0; x < len(msg.Additional); x++ {
		if msg.Additional[x].Type == RecordTypeOPT {
			continue
		}
		if msg.Additional[x].TTL < n {
			msg.Additional[x].TTL = 0
		} else {
			msg.Additional[x].TTL -= n
		}
		libbytes.WriteUint32(msg.packet, uint(msg.Additional[x].idxTTL), msg.Additional[x].TTL)
	}
}

// String return the message representation as string.
func (msg *Message) String() string {
	var (
		b strings.Builder
		x int
	)

	fmt.Fprintf(&b, "{Header:%+v Question:%+v", msg.Header, msg.Question)

	b.WriteString(" Answer:[")
	for x = 0; x < len(msg.Answer); x++ {
		if x > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%+v", msg.Answer[x])
	}
	b.WriteString("]")

	b.WriteString(" Authority:[")
	for x = 0; x < len(msg.Authority); x++ {
		if x > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%+v", msg.Authority[x])
	}
	b.WriteString("]")

	b.WriteString(" Additional:[")
	for x = 0; x < len(msg.Additional); x++ {
		if x > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%+v", msg.Additional[x])
	}
	b.WriteString("]}")

	return b.String()
}

// UnpackHeaderQuestion extract only DNS header and question from message
// packet.  This method assume that message.packet already set to DNS raw
// message.
func (msg *Message) UnpackHeaderQuestion() (err error) {
	if len(msg.packet) <= sectionHeaderSize {
		return errors.New(`UnpackHeaderQuestion: missing question`)
	}

	msg.Header.unpack(msg.packet)

	err = msg.Question.unpack(msg.packet[sectionHeaderSize:])
	if err != nil {
		return err
	}

	return nil
}
