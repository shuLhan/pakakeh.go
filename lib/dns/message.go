// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
	libnet "github.com/shuLhan/share/lib/net"
	"github.com/shuLhan/share/lib/reflect"
)

//
// Message represent a DNS message.
//
// All communications inside of the domain protocol are carried in a single
// format called a message.  The top level format of message is divided
// into 5 sections (some of which are empty in certain cases) shown below:
//
//     +---------------------+
//     |        Header       |
//     +---------------------+
//     |       Question      | the question for the name server
//     +---------------------+
//     |        Answer       | RRs answering the question
//     +---------------------+
//     |      Authority      | RRs pointing toward an authority
//     +---------------------+
//     |      Additional     | RRs holding additional information
//     +---------------------+
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
//
type Message struct {
	Header     SectionHeader
	Question   SectionQuestion
	Answer     []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord

	// Slice that hold the result of packing the message or original
	// message from unpacking.
	Packet []byte

	// offset of curret packet when packing, equal to len(Packet).
	off uint16

	// Mapping between name and their offset for message compression.
	dnameOff map[string]uint16
	dname    string
}

//
// NewMessage create, initialize, and return new message.
//
func NewMessage() *Message {
	return &Message{
		Header: SectionHeader{
			IsQuery: true,
			IsRD:    true,
			QDCount: 1,
		},
		Question: SectionQuestion{
			Type:  QueryTypeA,
			Class: QueryClassIN,
		},
		dnameOff: make(map[string]uint16),
	}
}

//
// NewMessageAddress create new DNS message for hostname that contains one or
// more A or AAAA addresses.
// The addresses must be all IPv4 or IPv6, the first address define the query
// type.
// If hname is not valid hostname or one of the address is not valid IP
// address it will return nil.
//
func NewMessageAddress(hname []byte, addresses [][]byte) (msg *Message) {
	if !libnet.IsHostnameValid(hname, false) {
		return nil
	}
	if len(addresses) == 0 {
		return nil
	}

	addr := addresses[0]
	qtype := getQueryTypeFromAddress(addr)
	if qtype == 0 {
		return nil
	}

	ascii.ToLower(&hname)

	rr := ResourceRecord{
		Name:  string(hname),
		Type:  qtype,
		Class: QueryClassIN,
		TTL:   defaultTTL,
		Value: string(addr),
	}

	msg = &Message{
		Header: SectionHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: SectionQuestion{
			Name:  string(hname),
			Type:  qtype,
			Class: QueryClassIN,
		},
		Answer: []ResourceRecord{rr},
	}

	for _, addr := range addresses[1:] {
		qtype = getQueryTypeFromAddress(addr)
		if qtype == 0 {
			continue
		}
		if qtype != msg.Question.Type {
			continue
		}
		msg.Answer = append(msg.Answer, ResourceRecord{
			Name:  string(hname),
			Type:  qtype,
			Class: QueryClassIN,
			TTL:   defaultTTL,
			Value: string(addr),
		})
		msg.Header.ANCount++
	}

	_, err := msg.Pack()
	if err != nil {
		return nil
	}

	return msg
}

//
// NewMessageFromRR create new message with one RR as an answer.
//
func NewMessageFromRR(rr *ResourceRecord) (msg *Message, err error) {
	msg = &Message{
		Header: SectionHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: SectionQuestion{
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

//
// getQueryTypeFromAddress return QueryTypeA or QueryTypeAAAA if addr is valid
// IPv4 or IPv6 address, otherwise it will return 0.
//
func getQueryTypeFromAddress(addr []byte) (qtype uint16) {
	ip := net.ParseIP(string(addr))
	if ip == nil {
		return 0
	}

	qtype = QueryTypeA
	for x := 0; x < len(addr); x++ {
		if addr[x] == ':' {
			qtype = QueryTypeAAAA
			break
		}
	}

	return qtype
}

//
// AddRR to the Answer field and re-pack it again.
//
func (msg *Message) AddRR(rr *ResourceRecord) (err error) {
	switch rr.Type {
	case QueryTypeSOA, QueryTypePTR:
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

//
// FilterAnswers return resource record in Answer that match only with
// specific query type.
//
func (msg *Message) FilterAnswers(t uint16) (answers []ResourceRecord) {
	for _, rr := range msg.Answer {
		if rr.Type == t {
			answers = append(answers, rr)
		}
	}
	return
}

func (msg *Message) compress() bool {
	off, ok := msg.dnameOff[msg.dname]
	if ok {
		msg.Packet = append(msg.Packet, maskPointer|byte(off>>8))
		msg.Packet = append(msg.Packet, byte(off))
		msg.off += 2
		return true
	}
	return false
}

//
// packDomainName convert string of domain-name into DNS domain-name format.
//
func (msg *Message) packDomainName(dname []byte, doCompress bool) (n int) {
	var (
		ok bool
		d  int
	)

	ascii.ToLower(&dname)
	msg.dname = string(dname)

	if doCompress {
		ok = msg.compress()
		if ok {
			return 2
		}
	}

	count := byte(0)
	msg.Packet = append(msg.Packet, 0)
	msg.dnameOff[msg.dname] = msg.off

	for x := 0; x < len(dname); x++ {
		c := dname[x]

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
				d, _ = strconv.Atoi(string(dname[x : x+3]))
				c = byte(d)
				if c >= 'A' && c <= 'Z' {
					c += 32
				}
				x += 2
			}
			msg.Packet = append(msg.Packet, c)
			count++
			continue
		}
		if c == '.' {
			// Skip name that prefixed with '.', e.g.
			// '...test.com'
			if count == 0 {
				continue
			}

			msg.Packet[msg.off] = count

			msg.dname = string(dname[x+1:])
			msg.off += uint16(count + 1)
			n += int(count + 1)

			if doCompress {
				ok = msg.compress()
				if ok {
					n += 2
					return n
				}
			}

			count = 0
			msg.Packet = append(msg.Packet, 0)
			msg.dnameOff[msg.dname] = msg.off

			if x+1 == len(dname) {
				return n
			}

			continue
		}

		msg.Packet = append(msg.Packet, c)
		count++
	}
	if count > 0 {
		msg.Packet[msg.off] = count
		msg.off += uint16(count + 1)
		n += int(count + 1)
	}
	if len(dname) > 0 {
		msg.Packet = append(msg.Packet, 0)
		msg.off++
		n++
	}

	return n
}

func (msg *Message) packQuestion() {
	msg.packDomainName([]byte(msg.Question.Name), false)
	libbytes.AppendUint16(&msg.Packet, msg.Question.Type)
	libbytes.AppendUint16(&msg.Packet, msg.Question.Class)
	msg.off += 4
}

func (msg *Message) packRR(rr *ResourceRecord) {
	var (
		rrOPT *RDataOPT
	)

	if rr.Type == QueryTypeOPT {
		// MUST be 0 (root domain).
		msg.Packet = append(msg.Packet, 0)
		rrOPT, _ = rr.Value.(*RDataOPT)
	} else {
		msg.packDomainName([]byte(rr.Name), true)
	}

	libbytes.AppendUint16(&msg.Packet, rr.Type)
	libbytes.AppendUint16(&msg.Packet, rr.Class)
	msg.off += 4

	if rr.Type == QueryTypeOPT {
		rr.TTL = 0

		// Pack extended code and version to TTL
		rr.TTL = uint32(rrOPT.ExtRCode) << 24
		rr.TTL |= (uint32(rrOPT.Version) << 16)

		if rrOPT.DO {
			rr.TTL |= maskOPTDO
		}
	}

	rr.offTTL = uint(msg.off)
	libbytes.AppendUint32(&msg.Packet, rr.TTL)
	msg.off += 4

	msg.packRData(rr)
}

func (msg *Message) packRData(rr *ResourceRecord) {
	switch rr.Type {
	case QueryTypeA:
		msg.packA(rr)
	case QueryTypeNS:
		msg.packTextAsDomain(rr)
	case QueryTypeMD:
		// obsolete
	case QueryTypeMF:
		// obsolete
	case QueryTypeCNAME:
		msg.packTextAsDomain(rr)
	case QueryTypeSOA:
		msg.packSOA(rr)
	case QueryTypeMB:
		msg.packTextAsDomain(rr)
	case QueryTypeMG:
		msg.packTextAsDomain(rr)
	case QueryTypeMR:
		msg.packTextAsDomain(rr)
	case QueryTypeNULL:
		msg.packTextAsDomain(rr)
	case QueryTypeWKS:
		msg.packWKS(rr)
	case QueryTypePTR:
		msg.packTextAsDomain(rr)
	case QueryTypeHINFO:
		msg.packHINFO(rr)
	case QueryTypeMINFO:
		msg.packMINFO(rr)
	case QueryTypeMX:
		msg.packMX(rr)
	case QueryTypeTXT:
		msg.packTXT(rr)
	case QueryTypeSRV:
		msg.packSRV(rr)
	case QueryTypeAAAA:
		msg.packAAAA(rr)
	case QueryTypeOPT:
		msg.packOPT(rr)
	}
}

func (msg *Message) packA(rr *ResourceRecord) {
	libbytes.AppendUint16(&msg.Packet, rdataIPv4Size)
	msg.off += 2

	rrText, _ := rr.Value.(string)

	ip := net.ParseIP(rrText)
	if ip == nil {
		msg.Packet = append(msg.Packet, rrText[:rdataIPv4Size]...)
	} else {
		ipv4 := ip.To4()
		if ipv4 == nil {
			msg.Packet = append(msg.Packet, ip[:rdataIPv4Size]...)
		} else {
			msg.Packet = append(msg.Packet, ipv4...)
		}
	}

	msg.off += rdataIPv4Size
}

func (msg *Message) packTextAsDomain(rr *ResourceRecord) {
	// Reserve two octets for rdlength
	libbytes.AppendUint16(&msg.Packet, 0)
	off := uint(msg.off)
	msg.off += 2

	rrText, _ := rr.Value.(string)

	n := msg.packDomainName([]byte(rrText), true)
	libbytes.WriteUint16(&msg.Packet, off, uint16(n))
}

func (msg *Message) packSOA(rr *ResourceRecord) {
	// Reserve two octets for rdlength.
	libbytes.AppendUint16(&msg.Packet, 0)
	off := uint(msg.off)
	msg.off += 2

	rrSOA, _ := rr.Value.(*RDataSOA)

	n := msg.packDomainName([]byte(rrSOA.MName), true)
	n += msg.packDomainName([]byte(rrSOA.RName), true)

	libbytes.AppendUint32(&msg.Packet, rrSOA.Serial)
	libbytes.AppendInt32(&msg.Packet, rrSOA.Refresh)
	libbytes.AppendInt32(&msg.Packet, rrSOA.Retry)
	libbytes.AppendInt32(&msg.Packet, rrSOA.Expire)
	libbytes.AppendUint32(&msg.Packet, rrSOA.Minimum)

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n+20))
	msg.off += uint16(n + 20)
}

func (msg *Message) packWKS(rr *ResourceRecord) {
	rrWKS, _ := rr.Value.(*RDataWKS)

	// Write rdlength.
	n := uint16(5 + len(rrWKS.BitMap))
	libbytes.AppendUint16(&msg.Packet, n)
	msg.off += 2

	msg.Packet = append(msg.Packet, rrWKS.Address[:4]...)
	msg.Packet = append(msg.Packet, rrWKS.Protocol)
	msg.Packet = append(msg.Packet, rrWKS.BitMap...)
	msg.off += n
}

func (msg *Message) packHINFO(rr *ResourceRecord) {
	rrHInfo, _ := rr.Value.(*RDataHINFO)

	// Write rdlength.
	n := len(rrHInfo.CPU)
	n += len(rrHInfo.OS)
	libbytes.AppendUint16(&msg.Packet, uint16(n))
	msg.off += 2
	msg.Packet = append(msg.Packet, rrHInfo.CPU...)
	msg.Packet = append(msg.Packet, rrHInfo.OS...)
	msg.off += uint16(n)
}

func (msg *Message) packMINFO(rr *ResourceRecord) {
	rrMInfo, _ := rr.Value.(*RDataMINFO)

	// Reserve two octets for rdlength.
	off := uint(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	n := msg.packDomainName([]byte(rrMInfo.RMailBox), true)
	n += msg.packDomainName([]byte(rrMInfo.EmailBox), true)

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n))
}

func (msg *Message) packMX(rr *ResourceRecord) {
	rrMX, _ := rr.Value.(*RDataMX)

	// Reserve two octets for rdlength.
	off := uint(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	libbytes.AppendInt16(&msg.Packet, rrMX.Preference)
	msg.off += 2

	n := msg.packDomainName([]byte(rrMX.Exchange), true)

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n+2))
}

func (msg *Message) packTXT(rr *ResourceRecord) {
	rrText, _ := rr.Value.(string)

	n := uint16(len(rrText))
	libbytes.AppendUint16(&msg.Packet, n+1)
	msg.off += 2

	msg.Packet = append(msg.Packet, byte(n))
	msg.Packet = append(msg.Packet, rrText...)
	msg.off += n
}

func (msg *Message) packSRV(rr *ResourceRecord) {
	rrSRV, _ := rr.Value.(*RDataSRV)

	// Reserve two octets for rdlength
	off := uint(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	libbytes.AppendUint16(&msg.Packet, rrSRV.Priority)
	msg.off += 2
	libbytes.AppendUint16(&msg.Packet, rrSRV.Weight)
	msg.off += 2
	libbytes.AppendUint16(&msg.Packet, rrSRV.Port)
	msg.off += 2

	n := msg.packDomainName([]byte(rrSRV.Target), false) + 6

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n))
}

func (msg *Message) packAAAA(rr *ResourceRecord) {
	rrText, _ := rr.Value.(string)

	libbytes.AppendUint16(&msg.Packet, rdataIPv6Size)
	msg.off += 2

	ip := net.ParseIP(rrText)
	if ip == nil {
		msg.Packet = append(msg.Packet, rrText[:rdataIPv6Size]...)
	} else {
		msg.Packet = append(msg.Packet, ip...)
	}

	msg.off += rdataIPv6Size
	msg.off += rdataIPv6Size
}

func (msg *Message) packOPT(rr *ResourceRecord) {
	rrOPT, _ := rr.Value.(*RDataOPT)

	// Reserve two octets for rdlength.
	off := uint(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	if rrOPT.Length == 0 {
		return
	}

	// Pack OPT rdata
	libbytes.AppendUint16(&msg.Packet, rrOPT.Code)

	// Values of less than 512 bytes MUST be treated as equal to 512
	// bytes (RFC6891 P11).
	if rrOPT.Length < 512 {
		libbytes.AppendUint16(&msg.Packet, 512)
	} else {
		libbytes.AppendUint16(&msg.Packet, rrOPT.Length)
	}

	msg.Packet = append(msg.Packet, rrOPT.Data[:rrOPT.Length]...)

	// Write rdlength.
	n := 4 + rrOPT.Length
	libbytes.WriteUint16(&msg.Packet, off, n)
	msg.off += n
}

//
// Reset the message fields.
//
func (msg *Message) Reset() {
	msg.Header.Reset()
	msg.Question.Reset()

	msg.ResetRR()
	msg.Packet = append(msg.Packet[:0], make([]byte, maxUDPPacketSize)...)

	msg.dname = ""
	msg.off = 0
	msg.dnameOff = make(map[string]uint16)
}

//
// ResetRR free allocated resource records in message.  This function can be
// used to release some memory after message has been packed, but the raw
// packet may still be in use.
//
func (msg *Message) ResetRR() {
	msg.Answer = nil
	msg.Authority = nil
	msg.Additional = nil
}

//
// IsExpired will return true if at least one resource record in answers is
// expired, where their TTL value is equal to 0.
// As long as the answers RR exist and no TTL is 0, it will return false.
//
// If RR answers is empty, then the TTL on authority RR will be checked for
// zero.
//
// There is no check to be done on additional RR, since its may contain EDNS
// with zero TTL.
//
func (msg *Message) IsExpired() bool {
	for x := 0; x < len(msg.Answer); x++ {
		if msg.Answer[x].TTL == 0 {
			return true
		}
	}
	if len(msg.Answer) > 0 {
		return false
	}

	for x := 0; x < len(msg.Authority); x++ {
		if msg.Authority[x].TTL == 0 {
			return true
		}
	}

	return false
}

//
// Pack convert message into datagram packet.  The result of packing
// a message will be saved in Packet field and returned.
//
func (msg *Message) Pack() ([]byte, error) {
	msg.dnameOff = make(map[string]uint16)
	msg.Packet = msg.Packet[:0]

	msg.Header.ANCount = uint16(len(msg.Answer))
	msg.Header.NSCount = uint16(len(msg.Authority))
	msg.Header.ARCount = uint16(len(msg.Additional))

	header := msg.Header.pack()

	msg.Packet = append(msg.Packet, header...)
	msg.off = uint16(sectionHeaderSize)

	msg.packQuestion()

	if msg.Header.IsQuery {
		msg.dnameOff = nil
		return msg.Packet, nil
	}

	for x := 0; x < len(msg.Answer); x++ {
		msg.packRR(&msg.Answer[x])
	}
	for x := 0; x < len(msg.Authority); x++ {
		msg.packRR(&msg.Authority[x])
	}
	for x := 0; x < len(msg.Additional); x++ {
		msg.packRR(&msg.Additional[x])
	}

	msg.dnameOff = nil

	return msg.Packet, nil
}

//
// RemoveAnswer remove the RR from list of answer.
//
func (msg *Message) RemoveAnswer(rr *ResourceRecord) (err error) {
	for x, an := range msg.Answer {
		fmt.Printf("RemoveAnswer: %s == %s?\n", an.Value, rr.Value)
		if !reflect.IsEqual(an.Value, rr.Value) {
			continue
		}
		copy(msg.Answer[x:], msg.Answer[x+1:])
		msg.Answer = msg.Answer[:len(msg.Answer)-1]
		msg.Header.ANCount--
		_, err = msg.Pack()
		break
	}
	return err
}

//
// SetAuthorativeAnswer set the header authoritative answer to true (1) or
// false (0).
//
func (msg *Message) SetAuthorativeAnswer(isAA bool) {
	msg.Header.IsAA = isAA
	if len(msg.Packet) > 2 {
		if isAA {
			msg.Packet[2] |= headerIsAA
		} else {
			msg.Packet[2] = (msg.Packet[2] & 0xFB)
		}
	}
}

//
// SetID in section header and in packet.
//
func (msg *Message) SetID(id uint16) {
	msg.Header.ID = id
	if len(msg.Packet) > 2 {
		libbytes.WriteUint16(&msg.Packet, 0, id)
	}
}

//
// SetQuery set the message as query (0) or as response (1) in header and in
// packet.
// Setting the message as query will also turning off AA, TC, and RA flags.
//
func (msg *Message) SetQuery(isQuery bool) {
	msg.Header.IsQuery = isQuery
	if len(msg.Packet) > 3 {
		if isQuery {
			// Turn off query, authoritative answer, and truncated
			// flags.
			msg.Packet[2] &= 0x71
			// Turn off recursion available flag.
			msg.Packet[3] &= 0x7F
		} else {
			msg.Packet[2] |= headerIsResponse
		}
	}
}

//
// SetRecursionDesired set the message to allow recursion (true=1) or
// not (false=0) in header and packet.
//
func (msg *Message) SetRecursionDesired(isRD bool) {
	msg.Header.IsRD = isRD
	if len(msg.Packet) > 2 {
		if isRD {
			msg.Packet[2] |= headerIsRD
		} else {
			msg.Packet[2] &= 0xFE
		}
	}
}

//
// SetResponseCode in message header and in packet.
//
func (msg *Message) SetResponseCode(code ResponseCode) {
	msg.Header.RCode = code
	if len(msg.Packet) > 3 {
		if code == RCodeOK {
			msg.Packet[3] &= 0xF0
		} else {
			msg.Packet[3] |= (0x0F & byte(code))
		}
	}
}

//
// SubTTL subtract TTL in each resource records and in packet by n seconds.
// If TTL is less than n, it will set to 0.
//
func (msg *Message) SubTTL(n uint32) {
	for x := 0; x < len(msg.Answer); x++ {
		if msg.Answer[x].TTL < n {
			msg.Answer[x].TTL = 0
		} else {
			msg.Answer[x].TTL -= n
		}
		libbytes.WriteUint32(&msg.Packet, msg.Answer[x].offTTL,
			msg.Answer[x].TTL)
	}
	for x := 0; x < len(msg.Authority); x++ {
		if msg.Authority[x].TTL < n {
			msg.Authority[x].TTL = 0
		} else {
			msg.Authority[x].TTL -= n
		}
		libbytes.WriteUint32(&msg.Packet, msg.Authority[x].offTTL,
			msg.Authority[x].TTL)
	}
	for x := 0; x < len(msg.Additional); x++ {
		if msg.Additional[x].Type == QueryTypeOPT {
			continue
		}
		if msg.Additional[x].TTL < n {
			msg.Additional[x].TTL = 0
		} else {
			msg.Additional[x].TTL -= n
		}
		libbytes.WriteUint32(&msg.Packet, msg.Additional[x].offTTL,
			msg.Additional[x].TTL)
	}
}

//
// String return the message representation as string.
//
func (msg *Message) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{Header:%+v Question:%+v", msg.Header, msg.Question)

	b.WriteString(" Answer:[")
	for x := 0; x < len(msg.Answer); x++ {
		if x > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%+v", msg.Answer[x])
	}
	b.WriteString("]")

	b.WriteString(" Authority:[")
	for x := 0; x < len(msg.Authority); x++ {
		if x > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%+v", msg.Authority[x])
	}
	b.WriteString("]")

	b.WriteString(" Additional:[")
	for x := 0; x < len(msg.Additional); x++ {
		if x > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%+v", msg.Additional[x])
	}
	b.WriteString("]}")

	return b.String()
}

//
// Unpack the packet to fill the message fields.
//
func (msg *Message) Unpack() (err error) {
	err = msg.UnpackHeaderQuestion()
	if err != nil {
		return err
	}

	startIdx := uint(sectionHeaderSize + msg.Question.size())

	var x uint16
	for ; x < msg.Header.ANCount; x++ {
		rr := ResourceRecord{}

		startIdx, err = rr.unpack(msg.Packet, startIdx)
		if err != nil {
			return err
		}

		msg.Answer = append(msg.Answer, rr)
	}

	if debug.Value >= 3 {
		log.Printf("msg.Answer: %+v\n", msg.Answer)
	}

	for x = 0; x < msg.Header.NSCount; x++ {
		rr := ResourceRecord{}

		startIdx, err = rr.unpack(msg.Packet, startIdx)
		if err != nil {
			return err
		}
		msg.Authority = append(msg.Authority, rr)
	}

	if debug.Value >= 3 {
		log.Printf("msg.Authority: %+v\n", msg.Authority)
	}

	for x = 0; x < msg.Header.ARCount; x++ {
		rr := ResourceRecord{}

		startIdx, err = rr.unpack(msg.Packet, startIdx)
		if err != nil {
			return err
		}

		msg.Additional = append(msg.Additional, rr)
	}

	if debug.Value >= 3 {
		log.Printf("msg.Additional: %+v\n", msg.Additional)
	}

	return nil
}

//
// UnpackHeaderQuestion extract only DNS header and question from message
// packet.  This method assume that message.Packet already set to DNS raw
// message.
//
func (msg *Message) UnpackHeaderQuestion() (err error) {
	msg.Header.unpack(msg.Packet)

	if debug.Value >= 3 {
		log.Printf("msg.Header: %+v\n", msg.Header)
	}

	if len(msg.Packet) <= sectionHeaderSize {
		return fmt.Errorf("Message.UnpackHeaderQuestion: missing question")
	}

	err = msg.Question.unpack(msg.Packet[sectionHeaderSize:])
	if err != nil {
		return err
	}

	if debug.Value >= 3 {
		log.Printf("msg.Question: %s\n", msg.Question.String())
	}

	return nil
}
