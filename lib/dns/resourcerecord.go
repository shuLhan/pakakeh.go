// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"log"

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
	Text  *RDataText
	SOA   *RDataSOA
	WKS   *RDataWKS
	HInfo *RDataHINFO
	MInfo *RDataMINFO
	MX    *RDataMX
	OPT   *RDataOPT

	off uint
}

//
// RData will return slice of bytes, the pointer that hold specific record
// data, or nil for obsolete type.
//
// For RR with type A, NS, CNAME, MB, MG, NULL, PTR, TXT or AAAA, it will
// return it as slice of bytes.
//
// For RR with type SOA, WKS, HINFO, MINFO, MX, or OPT it will return pointer
// to specific record type.
//
// For RR with obsolete type (MD or MF) it will return nil.
//
func (rr *ResourceRecord) RData() interface{} {
	switch rr.Type {
	case QueryTypeA:
		return rr.Text.v
	case QueryTypeNS:
		return rr.Text.v
	case QueryTypeMD:
		return nil
	case QueryTypeMF:
		return nil
	case QueryTypeCNAME:
		return rr.Text.v
	case QueryTypeSOA:
		return rr.SOA
	case QueryTypeMB:
		return rr.Text.v
	case QueryTypeMG:
		return rr.Text.v
	case QueryTypeNULL:
		return rr.Text.v
	case QueryTypeWKS:
		return rr.WKS
	case QueryTypePTR:
		return rr.Text.v
	case QueryTypeHINFO:
		return rr.HInfo
	case QueryTypeMINFO:
		return rr.MInfo
	case QueryTypeMX:
		return rr.MX
	case QueryTypeTXT:
		return rr.Text.v
	case QueryTypeAAAA:
		return rr.Text.v
	case QueryTypeOPT:
		return rr.OPT
	}
	return nil
}

//
// Reset the resource record fields to zero values.
//
func (rr *ResourceRecord) Reset() {
	rr.Name = rr.Name[:0]
	rr.Type = QueryTypeZERO
	rr.Class = QueryClassZERO
	rr.TTL = 0
	rr.rdlen = 0
	rr.rdata = rr.rdata[:0]
	rr.Text.v = rr.Text.v[:0]
	rr.SOA = nil
	rr.WKS = nil
	rr.HInfo = nil
	rr.MInfo = nil
	rr.MX = nil
	rr.OPT = nil
	rr.off = 0
}

func (rr *ResourceRecord) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "{Name:%s Type:%d Class:%d TTL:%d rdlen:%d}",
		rr.Name, rr.Type, rr.Class, rr.TTL, rr.rdlen)

	return buf.String()
}

//
// Unpack the DNS resource record from DNS packet start from index `startIdx`.
//
func (rr *ResourceRecord) Unpack(packet []byte, startIdx uint) (x uint, err error) {
	x = startIdx

	err = rr.unpackDomainName(&rr.Name, packet, x)
	if err != nil {
		return
	}
	if rr.off > 0 {
		x = rr.off + 1
	} else {
		if len(rr.Name) == 0 {
			x++
		} else {
			x = x + uint(len(rr.Name)+2)
		}
	}

	rr.Type = libbytes.ReadUint16(packet, x)
	x += 2
	rr.Class = uint16(libbytes.ReadUint16(packet, x))
	x += 2
	rr.TTL = libbytes.ReadUint32(packet, x)
	x += 4
	rr.rdlen = libbytes.ReadUint16(packet, x)
	x += 2

	rr.rdata = append(rr.rdata, packet[x:x+uint(rr.rdlen)]...)

	err = rr.unpackRData(packet, x)

	x = x + uint(rr.rdlen)

	return
}

func (rr *ResourceRecord) unpackDomainName(out *[]byte, packet []byte, x uint) error {
	count := packet[x]
	if count == 0 {
		return nil
	}
	if (packet[x] & maskPointer) == maskPointer {
		offset := uint16(packet[x]&maskOffset)<<8 | uint16(packet[x+1])

		if rr.off == 0 {
			rr.off = x + 1
		}

		err := rr.unpackDomainName(out, packet, uint(offset))
		return err
	}
	if count > maxLabelSize {
		return ErrLabelSizeLimit
	}
	if len(*out) > 0 {
		*out = append(*out, '.')
	}

	x++
	for y := byte(0); y < count; y++ {
		if packet[x] >= 'A' && packet[x] <= 'Z' {
			packet[x] += 32
		}
		*out = append(*out, packet[x])
		x++
	}

	err := rr.unpackDomainName(out, packet, x)

	return err
}

func (rr *ResourceRecord) unpackRData(packet []byte, startIdx uint) error {
	switch rr.Type {
	case QueryTypeA:
		if rr.rdlen != rdataIPv4Size || len(rr.rdata) != rdataIPv4Size {
			return ErrIPv4Length
		}
		rr.Text.v = append(rr.Text.v, rr.rdata...)

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
		return rr.unpackDomainName(&rr.Text.v, packet, startIdx)

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
		return rr.unpackDomainName(&rr.Text.v, packet, startIdx)

	case QueryTypeSOA:
		rr.SOA = new(RDataSOA)
		return rr.unpackSOA(packet, startIdx)

	case QueryTypeMB:
		return rr.unpackDomainName(&rr.Text.v, packet, startIdx)

	case QueryTypeMG:
		return rr.unpackDomainName(&rr.Text.v, packet, startIdx)

	// NULL records cause no additional section processing.
	// NULLs are used as placeholders in some experimental extensions of
	// the DNS.
	case QueryTypeNULL:
		endIdx := startIdx + uint(rr.rdlen)
		rr.Text.v = append(rr.Text.v, packet[startIdx:startIdx+endIdx]...)
		return nil

	case QueryTypeWKS:
		rr.WKS = new(RDataWKS)
		endIdx := startIdx + uint(rr.rdlen)
		return rr.WKS.UnmarshalBinary(packet[startIdx:endIdx])

	case QueryTypePTR:
		return rr.unpackDomainName(&rr.Text.v, packet, startIdx)

	case QueryTypeHINFO:
		rr.HInfo = new(RDataHINFO)
		endIdx := startIdx + uint(rr.rdlen)
		return rr.HInfo.UnmarshalBinary(packet[startIdx:endIdx])

	case QueryTypeMINFO:
		rr.MInfo = new(RDataMINFO)
		return rr.unpackMInfo(packet, startIdx)

	case QueryTypeMX:
		rr.MX = new(RDataMX)
		return rr.unpackMX(packet, startIdx)

	case QueryTypeTXT:
		endIdx := startIdx + uint(rr.rdlen)

		// The first byte of TXT is length.
		rr.Text.v = append(rr.Text.v, packet[startIdx+1:endIdx]...)

		return nil

	case QueryTypeAAAA:
		if rr.rdlen != rdataIPv6Size || len(rr.rdata) != rdataIPv6Size {
			return ErrIPv6Length
		}
		rr.Text.v = append(rr.Text.v, rr.rdata...)

	case QueryTypeOPT:
		rr.OPT = new(RDataOPT)
		return rr.unpackOPT(packet, startIdx)

	default:
		log.Printf("= Unknown query type: %d\n", rr.Type)
	}

	return nil
}

func (rr *ResourceRecord) unpackMInfo(packet []byte, startIdx uint) error {
	x := startIdx
	rr.off = 0

	err := rr.unpackDomainName(&rr.MInfo.RMailBox, packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x = x + uint(len(rr.MInfo.RMailBox)+2)
	}

	err = rr.unpackDomainName(&rr.MInfo.EmailBox, packet, x)
	if err != nil {
		return err
	}

	return nil
}

func (rr *ResourceRecord) unpackMX(packet []byte, startIdx uint) error {
	rr.MX.Preference = libbytes.ReadInt16(packet, startIdx)

	rr.off = 0
	err := rr.unpackDomainName(&rr.MX.Exchange, packet, startIdx+2)

	return err
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
	rr.OPT.Data = append(rr.OPT.Data, packet[x:endIdx]...)
	return nil
}

func (rr *ResourceRecord) unpackSOA(packet []byte, startIdx uint) error {
	x := startIdx
	rr.off = 0

	err := rr.unpackDomainName(&rr.SOA.MName, packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x = x + uint(len(rr.SOA.MName)+2)
	}

	err = rr.unpackDomainName(&rr.SOA.RName, packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x = x + uint(len(rr.SOA.RName)+2)
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
