// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"log"

	libbytes "github.com/shuLhan/share/lib/bytes"
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
	Header     *SectionHeader
	Question   *SectionQuestion
	Answer     []*ResourceRecord
	Authority  []*ResourceRecord
	Additional []*ResourceRecord

	// Slice that hold the result of packing the message or original
	// message from unpacking.
	Packet []byte

	// offset of curret packet when packing, equal to len(Packet).
	off uint16

	// Mapping between name and their offset for message compression.
	dnameOff map[string]uint16
	dname    string
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
func (msg *Message) packDomainName(dname []byte) (n int) {
	msg.dname = string(dname)
	ok := msg.compress()
	if ok {
		n = 2
		return
	}

	count := byte(0)
	msg.Packet = append(msg.Packet, 0)
	msg.dnameOff[msg.dname] = msg.off

	for x, c := range dname {
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

			ok = msg.compress()
			if ok {
				n += 2
				return
			}

			count = 0
			msg.Packet = append(msg.Packet, 0)
			msg.dnameOff[msg.dname] = msg.off

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

	return
}

func (msg *Message) packQuestion() {
	msg.packDomainName(msg.Question.Name)
	libbytes.AppendUint16(&msg.Packet, uint16(msg.Question.Type))
	libbytes.AppendUint16(&msg.Packet, uint16(msg.Question.Class))
	msg.off += 4
}

func (msg *Message) packRR(rr *ResourceRecord) {
	msg.packDomainName(rr.Name)
	libbytes.AppendUint16(&msg.Packet, uint16(rr.Type))
	libbytes.AppendUint16(&msg.Packet, uint16(rr.Class))
	msg.off += 4

	if rr.Type == QueryTypeOPT {
		rr.TTL = 0

		// Pack extended code and version to TTL
		rr.TTL = uint32(rr.OPT.ExtRCode) << 24
		rr.TTL = rr.TTL | uint32(rr.OPT.Version)<<16

		if rr.OPT.DO {
			rr.TTL = rr.TTL | maskOPTDO
		}
	}

	libbytes.AppendUint32(&msg.Packet, rr.TTL)
	msg.off += 4

	msg.packRData(rr)
}

func (msg *Message) packRData(rr *ResourceRecord) {
	switch rr.Type {
	case QueryTypeA:
		if len(rr.Text.v) >= rdataIPv4Size {
			libbytes.AppendUint16(&msg.Packet, rdataIPv4Size)
			msg.off += 2
			msg.Packet = append(msg.Packet, rr.Text.v[:rdataIPv4Size]...)
			msg.off += rdataIPv4Size
		}
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
	case QueryTypeAAAA:
		msg.packAAAA(rr)
	case QueryTypeOPT:
		msg.packOPT(rr)
	}
}

func (msg *Message) packTextAsDomain(rr *ResourceRecord) {
	// Reserve two octets for rdlength
	libbytes.AppendUint16(&msg.Packet, 0)
	off := int(msg.off)
	msg.off += 2

	n := msg.packDomainName(rr.Text.v)
	libbytes.WriteUint16(&msg.Packet, off, uint16(n))
}

func (msg *Message) packSOA(rr *ResourceRecord) {
	// Reserve two octets for rdlength.
	libbytes.AppendUint16(&msg.Packet, 0)
	off := int(msg.off)
	msg.off += 2

	n := msg.packDomainName(rr.SOA.MName)
	n += msg.packDomainName(rr.SOA.RName)

	libbytes.AppendUint32(&msg.Packet, rr.SOA.Serial)
	libbytes.AppendInt32(&msg.Packet, rr.SOA.Refresh)
	libbytes.AppendInt32(&msg.Packet, rr.SOA.Retry)
	libbytes.AppendInt32(&msg.Packet, rr.SOA.Expire)
	libbytes.AppendUint32(&msg.Packet, rr.SOA.Minimum)

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n+20))
	msg.off += uint16(n + 20)
}

func (msg *Message) packWKS(rr *ResourceRecord) {
	// Write rdlength.
	n := uint16(5 + len(rr.WKS.BitMap))
	libbytes.AppendUint16(&msg.Packet, n)
	msg.off += 2

	msg.Packet = append(msg.Packet, rr.WKS.Address[:4]...)
	msg.Packet = append(msg.Packet, rr.WKS.Protocol)
	msg.Packet = append(msg.Packet, rr.WKS.BitMap...)
	msg.off += n
}

func (msg *Message) packHINFO(rr *ResourceRecord) {
	// Write rdlength.
	n := len(rr.HInfo.CPU)
	n += len(rr.HInfo.OS)
	libbytes.AppendUint16(&msg.Packet, uint16(n))
	msg.off += 2
	msg.Packet = append(msg.Packet, rr.HInfo.CPU...)
	msg.Packet = append(msg.Packet, rr.HInfo.OS...)
	msg.off += uint16(n)
}

func (msg *Message) packMINFO(rr *ResourceRecord) {
	// Reserve two octets for rdlength.
	off := int(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	n := msg.packDomainName(rr.MInfo.RMailBox)
	n += msg.packDomainName(rr.MInfo.EmailBox)

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n))
}

func (msg *Message) packMX(rr *ResourceRecord) {
	// Reserve two octets for rdlength.
	off := int(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	libbytes.AppendInt16(&msg.Packet, rr.MX.Preference)
	msg.off += 2

	n := msg.packDomainName(rr.MX.Exchange)

	// Write rdlength.
	libbytes.WriteUint16(&msg.Packet, off, uint16(n+2))
}

func (msg *Message) packTXT(rr *ResourceRecord) {
	n := uint16(len(rr.Text.v))
	libbytes.AppendUint16(&msg.Packet, n+1)
	msg.off += 2

	msg.Packet = append(msg.Packet, byte(n))
	msg.Packet = append(msg.Packet, rr.Text.v...)
	msg.off += n
}

func (msg *Message) packAAAA(rr *ResourceRecord) {
	libbytes.AppendUint16(&msg.Packet, rdataIPv6Size)
	msg.off += 2

	msg.Packet = append(msg.Packet, rr.Text.v[:rdataIPv6Size]...)
	msg.off += rdataIPv6Size
}

func (msg *Message) packOPT(rr *ResourceRecord) {
	// Reserve two octets for rdlength.
	off := int(msg.off)
	libbytes.AppendUint16(&msg.Packet, 0)
	msg.off += 2

	if rr.OPT.Length == 0 {
		return
	}

	// Pack OPT rdata
	libbytes.AppendUint16(&msg.Packet, rr.OPT.Code)
	libbytes.AppendUint16(&msg.Packet, rr.OPT.Length)
	msg.Packet = append(msg.Packet, rr.OPT.Data[:rr.OPT.Length]...)

	// Write rdlength.
	n := uint16(4 + rr.OPT.Length)
	libbytes.WriteUint16(&msg.Packet, off, uint16(n))
	msg.off += n
}

//
// Reset the message fields.
//
func (msg *Message) Reset() {
	msg.Header.Reset()
	msg.Question.Reset()

	for x := 0; x < len(msg.Answer); x++ {
		rrPool.Put(msg.Answer[x])
	}
	for x := 0; x < len(msg.Authority); x++ {
		rrPool.Put(msg.Authority[x])
	}
	for x := 0; x < len(msg.Additional); x++ {
		rrPool.Put(msg.Additional[x])
	}

	msg.Answer = msg.Answer[:0]
	msg.Authority = msg.Authority[:0]
	msg.Additional = msg.Additional[:0]
	msg.Packet = msg.Packet[:0]

	msg.dname = ""
	msg.off = 0
	msg.dnameOff = make(map[string]uint16)
}

//
// MarshalBinary convert message into datagram packet.  The result of packing
// a message will be saved in Packet field and returned.
//
func (msg *Message) MarshalBinary() ([]byte, error) {
	if msg.dnameOff == nil {
		msg.dnameOff = make(map[string]uint16)
	}

	msg.Packet = msg.Packet[:0]

	msg.Header.ANCount = uint16(len(msg.Answer))
	msg.Header.NSCount = uint16(len(msg.Authority))
	msg.Header.ARCount = uint16(len(msg.Additional))

	header, err := msg.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}

	msg.Packet = append(msg.Packet, header...)
	msg.off = uint16(sectionHeaderSize)

	msg.packQuestion()

	if msg.Header.IsQuery {
		return msg.Packet, nil
	}

	for x := 0; x < len(msg.Answer); x++ {
		msg.packRR(msg.Answer[x])
	}
	for x := 0; x < len(msg.Authority); x++ {
		msg.packRR(msg.Authority[x])
	}
	for x := 0; x < len(msg.Additional); x++ {
		msg.packRR(msg.Additional[x])
	}

	return msg.Packet, nil
}

//
// UnmarshalBinary unpack the packet to fill the message fields.
//
func (msg *Message) UnmarshalBinary(packet []byte) error {
	msg.Packet = packet

	_ = msg.Header.UnmarshalBinary(packet)

	if debugLevel >= 1 {
		log.Printf("msg.Header: %+v\n", msg.Header)
	}

	if len(packet) <= sectionHeaderSize {
		return nil
	}

	err := msg.Question.UnmarshalBinary(packet[sectionHeaderSize:])
	if err != nil {
		return err
	}

	if debugLevel >= 1 {
		log.Printf("msg.Question: %s\n", msg.Question)
	}

	startIdx := sectionHeaderSize + msg.Question.Size()

	var x uint16
	for ; x < msg.Header.ANCount; x++ {
		rr := rrPool.Get().(*ResourceRecord)
		rr.Reset()
		startIdx, err = rr.Unpack(packet, startIdx)
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
	}

	if debugLevel >= 1 {
		log.Printf("msg.Answer: %+v\n", msg.Answer)
	}

	for x = 0; x < msg.Header.NSCount; x++ {
		rr := rrPool.Get().(*ResourceRecord)
		rr.Reset()
		startIdx, err = rr.Unpack(packet, startIdx)
		if err != nil {
			return err
		}
		msg.Authority = append(msg.Authority, rr)
	}

	if debugLevel >= 1 {
		log.Printf("msg.Authority: %+v\n", msg.Authority)
	}

	for x = 0; x < msg.Header.ARCount; x++ {
		rr := rrPool.Get().(*ResourceRecord)
		rr.Reset()
		startIdx, err = rr.Unpack(packet, startIdx)
		if err != nil {
			return err
		}
		msg.Additional = append(msg.Additional, rr)
	}

	if debugLevel >= 1 {
		log.Printf("msg.Additional: %+v\n", msg.Additional)
	}

	return nil
}
