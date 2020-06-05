// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// ResourceRecord The answer, authority, and additional sections all share the
// same format: a variable number of resource records, where the number of
// records is specified in the corresponding count field in the header.  Each
// resource record has the following format:
//
type ResourceRecord struct {
	// A domain name to which this resource record pertains.
	Name []byte

	// Two octets containing one of the RR type codes.  This field
	// specifies the meaning of the data in the RDATA field.
	Type uint16

	// Two octets which specify the class of the data in the RDATA field.
	Class uint16

	// A 32 bit unsigned integer that specifies the time interval (in
	// seconds) that the resource record may be cached before it should be
	// discarded.  Zero values are interpreted to mean that the RR can
	// only be used for the transaction in progress, and should not be
	// cached.
	TTL uint32

	// An unsigned 16 bit integer that specifies the length in octets of
	// the RDATA field.
	rdlen uint16

	// A variable length string of octets that describes the resource.
	// The format of this information varies according to the TYPE and
	// CLASS of the resource record.  For example, if the TYPE is A
	// and the CLASS is IN, the RDATA field is a 4 octet ARPA Internet
	// address.
	rdata []byte

	// Text represent A, NS, CNAME, MB, MG, NULL, PTR, and TXT.
	Text  []byte
	SOA   *RDataSOA
	WKS   *RDataWKS
	HInfo *RDataHINFO
	MInfo *RDataMINFO
	MX    *RDataMX
	OPT   *RDataOPT
	SRV   *RDataSRV

	off    uint
	offTTL uint
}

//
// NewResourceRecord create and initialize new ResourceRecord.
//
func NewResourceRecord() *ResourceRecord {
	return &ResourceRecord{
		Name:  make([]byte, 0),
		rdata: make([]byte, 0),
	}
}

//
// RData will return slice of bytes, the pointer that hold specific record
// data, or nil for obsolete type.
//
// For RR with type A, NS, CNAME, MB, MG, MR, NULL, PTR, TXT or AAAA, it will
// return it as slice of bytes.
//
// For RR with type SOA, WKS, HINFO, MINFO, MX, OPT, or SRV it will return
// pointer to specific record type.
//
// For RR with obsolete type (MD or MF) it will return nil.
//
func (rr *ResourceRecord) RData() interface{} {
	switch rr.Type {
	case QueryTypeA:
		return rr.Text
	case QueryTypeNS:
		return rr.Text
	case QueryTypeMD:
		return nil
	case QueryTypeMF:
		return nil
	case QueryTypeCNAME:
		return rr.Text
	case QueryTypeSOA:
		return rr.SOA
	case QueryTypeMB:
		return rr.Text
	case QueryTypeMG:
		return rr.Text
	case QueryTypeMR:
		return rr.Text
	case QueryTypeNULL:
		return rr.Text
	case QueryTypeWKS:
		return rr.WKS
	case QueryTypePTR:
		return rr.Text
	case QueryTypeHINFO:
		return rr.HInfo
	case QueryTypeMINFO:
		return rr.MInfo
	case QueryTypeMX:
		return rr.MX
	case QueryTypeTXT:
		return rr.Text
	case QueryTypeAAAA:
		return rr.Text
	case QueryTypeSRV:
		return rr.SRV
	case QueryTypeOPT:
		return rr.OPT
	}
	return nil
}

//
// String return the text representation of ResourceRecord for human.
//
func (rr *ResourceRecord) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "{Name:%s Type:%d Class:%d TTL:%d rdlen:%d}",
		rr.Name, rr.Type, rr.Class, rr.TTL, rr.rdlen)

	return buf.String()
}

//
// unpack the DNS resource record from DNS packet start from index `startIdx`.
//
func (rr *ResourceRecord) unpack(packet []byte, startIdx uint) (x uint, err error) {
	x = startIdx

	rr.Name, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return x, err
	}
	if rr.off > 0 {
		x = rr.off + 1
	} else {
		if len(rr.Name) == 0 {
			x++
		} else {
			x += uint(len(rr.Name) + 2)
		}
	}

	rr.Type = libbytes.ReadUint16(packet, x)
	x += 2
	rr.Class = libbytes.ReadUint16(packet, x)
	x += 2
	rr.offTTL = x
	rr.TTL = libbytes.ReadUint32(packet, x)
	x += 4
	rr.rdlen = libbytes.ReadUint16(packet, x)
	x += 2

	rr.rdata = append(rr.rdata, packet[x:x+uint(rr.rdlen)]...)

	err = rr.unpackRData(packet, x)

	x += uint(rr.rdlen)

	return x, err
}

func (rr *ResourceRecord) unpackDomainName(packet []byte, start uint) (
	out []byte, err error,
) {
	x := int(start)
	for x < len(packet) {
		count := packet[x]
		if count == 0 {
			break
		}
		if (packet[x] & maskPointer) == maskPointer {
			offset := uint16(packet[x]&maskOffset)<<8 | uint16(packet[x+1])

			if rr.off == 0 {
				rr.off = uint(x + 1)
			}
			x = int(offset)
			continue
		}
		if count > maxLabelSize {
			return nil, ErrLabelSizeLimit
		}
		if len(out) > 0 {
			out = append(out, '.')
		}

		x++
		for y := byte(0); y < count; y++ {
			if x >= len(packet) {
				break
			}
			if packet[x] >= 'A' && packet[x] <= 'Z' {
				packet[x] += 32
			}
			out = append(out, packet[x])
			x++
		}
	}
	return out, nil
}

func (rr *ResourceRecord) unpackRData(packet []byte, startIdx uint) (err error) {
	switch rr.Type {
	case QueryTypeA:
		return rr.unpackA()

	//
	// NS records cause both the usual additional section processing to
	// locate a type A record, and, when used in a referral, a special
	// search of the zone in which they reside for glue information.
	//
	// The NS RR states that the named host should be expected to have a
	// zone starting at owner name of the specified class.  Note that the
	// class may not indicate the protocol family which should be used to
	// communicate with the host, although it is typically a strong hint.
	// For example, hosts which are name servers for either Internet (IN)
	// or Hesiod (HS) class information are normally queried using IN
	// class protocols.
	//
	case QueryTypeNS:
		rr.Text, err = rr.unpackDomainName(packet, startIdx)
		return err

	// MD is obsolete.  See the definition of MX and [RFC-974] for details of
	// the new scheme.  The recommended policy for dealing with MD RRs found in
	// a master file is to reject them, or to convert them to MX RRs with a
	// preference of 0.
	case QueryTypeMD:

	// MF is obsolete.  See the definition of MX and [RFC-974] for details
	// ofw the new scheme.  The recommended policy for dealing with MD RRs
	// found in a master file is to reject them, or to convert them to MX
	// RRs with a preference of 10.
	case QueryTypeMF:

	// CNAME RRs cause no additional section processing, but name servers
	// may choose to restart the query at the canonical name in certain
	// cases.  See the description of name server logic in [RFC-1034] for
	// details.
	case QueryTypeCNAME:
		rr.Text, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeSOA:
		rr.SOA = new(RDataSOA)
		return rr.unpackSOA(packet, startIdx)

	case QueryTypeMB:
		rr.Text, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeMG:
		rr.Text, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeMR:
		rr.Text, err = rr.unpackDomainName(packet, startIdx)
		return err

	// NULL records cause no additional section processing.
	// NULLs are used as placeholders in some experimental extensions of
	// the DNS.
	case QueryTypeNULL:
		endIdx := startIdx + uint(rr.rdlen)
		rr.Text = packet[startIdx : startIdx+endIdx]
		return nil

	case QueryTypeWKS:
		rr.WKS = new(RDataWKS)
		endIdx := startIdx + uint(rr.rdlen)
		return rr.WKS.unpack(packet[startIdx:endIdx])

	case QueryTypePTR:
		rr.Text, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeHINFO:
		rr.HInfo = new(RDataHINFO)
		endIdx := startIdx + uint(rr.rdlen)
		return rr.HInfo.unpack(packet[startIdx:endIdx])

	case QueryTypeMINFO:
		rr.MInfo = new(RDataMINFO)
		return rr.unpackMInfo(packet, startIdx)

	case QueryTypeMX:
		rr.MX = new(RDataMX)
		return rr.unpackMX(packet, startIdx)

	case QueryTypeTXT:
		endIdx := startIdx + uint(rr.rdlen)

		// The first byte of TXT is length.
		rr.Text = packet[startIdx+1 : endIdx]

		return nil

	case QueryTypeAAAA:
		return rr.unpackAAAA()

	case QueryTypeSRV:
		rr.SRV = new(RDataSRV)
		return rr.unpackSRV(packet, startIdx)

	case QueryTypeOPT:
		rr.OPT = new(RDataOPT)
		return rr.unpackOPT(packet, startIdx)

	default:
		log.Printf("= Unknown query type: %d\n", rr.Type)
	}

	return nil
}

func (rr *ResourceRecord) unpackA() error {
	if rr.rdlen != rdataIPv4Size || len(rr.rdata) != rdataIPv4Size {
		return ErrIPv4Length
	}

	ip := net.IP(rr.rdata)
	rr.Text = []byte(ip.String())

	return nil
}

func (rr *ResourceRecord) unpackAAAA() error {
	if rr.rdlen != rdataIPv6Size || len(rr.rdata) != rdataIPv6Size {
		return ErrIPv6Length
	}

	ip := net.IP(rr.rdata)
	rr.Text = []byte(ip.String())

	return nil
}

func (rr *ResourceRecord) unpackMInfo(packet []byte, startIdx uint) (err error) {
	x := startIdx
	rr.off = 0

	rr.MInfo.RMailBox, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x += uint(len(rr.MInfo.RMailBox) + 2)
	}

	rr.MInfo.EmailBox, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}

	return nil
}

func (rr *ResourceRecord) unpackMX(packet []byte, startIdx uint) (err error) {
	rr.MX.Preference = libbytes.ReadInt16(packet, startIdx)

	rr.off = 0
	rr.MX.Exchange, err = rr.unpackDomainName(packet, startIdx+2)

	return err
}

func (rr *ResourceRecord) unpackSRV(packet []byte, x uint) (err error) {
	// Unpack service, proto, and name from RR.Name
	y := 0
	for ; y < len(rr.Name); y++ {
		if rr.Name[y] == '.' {
			break
		}
		rr.SRV.Service = append(rr.SRV.Service, rr.Name[y])
	}
	for y++; y < len(rr.Name); y++ {
		if rr.Name[y] == '.' {
			break
		}
		rr.SRV.Proto = append(rr.SRV.Proto, rr.Name[y])
	}
	for y++; y < len(rr.Name); y++ {
		rr.SRV.Name = append(rr.SRV.Name, rr.Name[y])
	}

	// Unpack RDATA
	rr.SRV.Priority = libbytes.ReadUint16(packet, x)
	x += 2
	rr.SRV.Weight = libbytes.ReadUint16(packet, x)
	x += 2
	rr.SRV.Port = libbytes.ReadUint16(packet, x)
	x += 2

	rr.SRV.Target, err = rr.unpackDomainName(packet, x)

	return
}

func (rr *ResourceRecord) unpackOPT(packet []byte, x uint) error {
	// Unpack extended RCODE and flags from TTL.
	rr.OPT.ExtRCode = byte(rr.TTL >> 24)
	rr.OPT.Version = byte(rr.TTL >> 16)

	if rr.TTL&maskOPTDO == maskOPTDO {
		rr.OPT.DO = true
	}

	if rr.rdlen == 0 {
		return nil
	}

	// Unpack the RDATA
	rr.OPT.Code = libbytes.ReadUint16(packet, x)
	x += 2
	rr.OPT.Length = libbytes.ReadUint16(packet, x)
	x += 2
	endIdx := x + uint(rr.rdlen)
	if int(endIdx) >= len(packet) {
		return errors.New("RR OPT length is out of range")
	}
	rr.OPT.Data = append(rr.OPT.Data, packet[x:endIdx]...)
	return nil
}

func (rr *ResourceRecord) unpackSOA(packet []byte, startIdx uint) (err error) {
	x := startIdx
	rr.off = 0

	rr.SOA.MName, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x += uint(len(rr.SOA.MName) + 2)
	}

	rr.SOA.RName, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x += uint(len(rr.SOA.RName) + 2)
	}

	rr.SOA.Serial = libbytes.ReadUint32(packet, x)
	x += 4
	rr.SOA.Refresh = libbytes.ReadInt32(packet, x)
	x += 4
	rr.SOA.Retry = libbytes.ReadInt32(packet, x)
	x += 4
	rr.SOA.Expire = libbytes.ReadInt32(packet, x)
	x += 4
	rr.SOA.Minimum = libbytes.ReadUint32(packet, x)

	return nil
}
