// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

const (
	maskPointer byte  = 0xC0
	maskOffset  byte  = 0x3F
	maskOPTDO   int32 = 0x00008000
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
	Type QueryType

	// Two octets which specify the class of the data in the RDATA field.
	Class QueryClass

	// A 32 bit unsigned integer that specifies the time interval (in
	// seconds) that the resource record may be cached before it should be
	// discarded.  Zero values are interpreted to mean that the RR can
	// only be used for the transaction in progress, and should not be
	// cached.
	TTL int32

	// An unsigned 16 bit integer that specifies the length in octets of
	// the RDATA field.
	RDLength uint16

	// A variable length string of octets that describes the resource.
	// The format of this information varies according to the TYPE and
	// CLASS of the resource record.  For example, if the TYPE is A
	// and the CLASS is IN, the RDATA field is a 4 octet ARPA Internet
	// address.
	rdata []byte

	// rdataText represent A, NS, CNAME, MB, MG, NULL, PTR, and TXT.
	rdataText *RDataText

	rdataSOA *RDataSOA

	// The WKS record is used to describe the well known services
	// supported by a particular protocol on a particular internet
	// address.
	rdataWKS *RDataWKS

	rdataHINFO *RDataHINFO
	rdataMINFO *RDataMINFO
	rdataMX    *RDataMX
	rdataOPT   *RDataOPT

	offsetIdx int
}

//
// RData will return slice of bytes, the pointer that hold specific record
// data, or nil for obsolete type.
//
// For RR with type A, NS, CNAME, MB, MG, NULL, PTR, or TXT it will return
// slice of bytes.
//
// For RR with type SOA, WKS, HINFO, MINFO, or MX it will return pointer to
// specific record type.
//
// For RR with absolute type (MD or MF) it will return nil.
//
func (rr *ResourceRecord) RData() interface{} {
	switch rr.Type {
	case QueryTypeA:
		return rr.rdataText.v
	case QueryTypeNS:
		return rr.rdataText.v
	case QueryTypeMD:
		return nil
	case QueryTypeMF:
		return nil
	case QueryTypeCNAME:
		return rr.rdataText.v
	case QueryTypeSOA:
		return rr.rdataSOA
	case QueryTypeMB:
		return rr.rdataText.v
	case QueryTypeMG:
		return rr.rdataText.v
	case QueryTypeNULL:
		return rr.rdataText.v
	case QueryTypeWKS:
		return rr.rdataWKS
	case QueryTypePTR:
		return rr.rdataText.v
	case QueryTypeHINFO:
		return rr.rdataHINFO
	case QueryTypeMINFO:
		return rr.rdataMINFO
	case QueryTypeMX:
		return rr.rdataMX
	case QueryTypeTXT:
		return rr.rdataText.v
	case QueryTypeOPT:
		return rr.rdataOPT
	}
	return nil
}

//
// Reset the resource record fields to zero values.
//
func (rr *ResourceRecord) Reset() {
	rr.Name = rr.Name[:0]
	rr.offsetIdx = 0
	rr.Type = QueryTypeZERO
	rr.Class = QueryClassZERO
	rr.TTL = 0
	rr.RDLength = 0
	rr.rdata = rr.rdata[:0]
	rr.rdataSOA = nil
	rr.rdataWKS = nil
	rr.rdataHINFO = nil
	rr.rdataMINFO = nil
	rr.rdataMX = nil
	rr.rdataOPT = nil
}

func (rr *ResourceRecord) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "{Name:%s Type:%d Class:%d TTL:%d RDLength:%d",
		rr.Name, rr.Type, rr.Class, rr.TTL, rr.RDLength)

	rdata := rr.RData()
	if rdata != nil {
		fmt.Fprintf(&buf, " rdata:%s}", rdata)
	} else {
		buf.WriteString(" rdata:nil}")
	}

	return buf.String()
}

//
// Unpack the DNS resource record from DNS packet start from index `startIdx`.
//
func (rr *ResourceRecord) Unpack(packet []byte, startIdx int) (int, error) {
	x := startIdx

	err := rr.unpackDomainName(&rr.Name, packet, x)
	if err != nil {
		return 0, err
	}
	if rr.offsetIdx > 0 {
		x = rr.offsetIdx + 1
	} else {
		if len(rr.Name) == 0 {
			x++
		} else {
			x = x + len(rr.Name) + 2
		}
	}

	rr.Type = QueryType(libbytes.ReadUint16(packet, x))
	x += 2
	rr.Class = QueryClass(libbytes.ReadUint16(packet, x))
	x += 2
	rr.TTL = libbytes.ReadInt32(packet, x)
	x += 4
	rr.RDLength = libbytes.ReadUint16(packet, x)
	x += 2

	rr.rdata = append(rr.rdata, packet[x:x+int(rr.RDLength)]...)

	rr.unpackRData(packet, x)

	startIdx = x + int(rr.RDLength)

	return startIdx, nil
}

func (rr *ResourceRecord) unpackDomainName(out *[]byte, packet []byte, x int) error {
	count := packet[x]
	if count == 0 {
		return nil
	}
	if (packet[x] & maskPointer) == maskPointer {
		offset := uint16(packet[x]&maskOffset)<<8 | uint16(packet[x+1])

		if rr.offsetIdx == 0 {
			rr.offsetIdx = x + 1
		}

		err := rr.unpackDomainName(out, packet, int(offset))
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
		*out = append(*out, packet[x])
		x++
	}

	err := rr.unpackDomainName(out, packet, x)

	return err
}

func (rr *ResourceRecord) unpackRData(packet []byte, startIdx int) error {
	switch rr.Type {
	case QueryTypeA:
		if rr.RDLength != rdataAddrSize || len(rr.rdata) != rdataAddrSize {
			return ErrRDataAddrLength
		}
		rr.rdataText = new(RDataText)
		rr.rdataText.v = append(rr.rdataText.v, rr.rdata...)

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
		rr.rdataText = new(RDataText)
		return rr.unpackDomainName(&rr.rdataText.v, packet, startIdx)

	// MD is obsolete.  See the definition of MX and [RFC-974] for details of
	// the new scheme.  The recommended policy for dealing with MD RRs found in
	// a master file is to reject them, or to convert them to MX RRs with a
	// preference of 0.
	case QueryTypeMD:
		return nil

	// MF is obsolete.  See the definition of MX and [RFC-974] for details
	// ofw the new scheme.  The recommended policy for dealing with MD RRs
	// found in a master file is to reject them, or to convert them to MX
	// RRs with a preference of 10.
	case QueryTypeMF:
		return nil

	// CNAME RRs cause no additional section processing, but name servers
	// may choose to restart the query at the canonical name in certain
	// cases.  See the description of name server logic in [RFC-1034] for
	// details.
	case QueryTypeCNAME:
		rr.rdataText = new(RDataText)
		return rr.unpackDomainName(&rr.rdataText.v, packet, startIdx)

	case QueryTypeSOA:
		rr.rdataSOA = new(RDataSOA)
		return rr.unpackRDataSOA(packet, startIdx)

	case QueryTypeMB:
		rr.rdataText = new(RDataText)
		return rr.unpackDomainName(&rr.rdataText.v, packet, startIdx)

	case QueryTypeMG:
		rr.rdataText = new(RDataText)
		return rr.unpackDomainName(&rr.rdataText.v, packet, startIdx)

	// NULL records cause no additional section processing.
	// NULLs are used as placeholders in some experimental extensions of
	// the DNS.
	case QueryTypeNULL:
		rr.rdataText = new(RDataText)
		endIdx := startIdx + int(rr.RDLength)
		rr.rdataText.v = append(rr.rdataText.v, packet[startIdx:startIdx+endIdx]...)
		return nil

	case QueryTypeWKS:
		rr.rdataWKS = new(RDataWKS)
		endIdx := startIdx + int(rr.RDLength)
		return rr.rdataWKS.UnmarshalBinary(packet[startIdx:endIdx])

	case QueryTypePTR:
		rr.rdataText = new(RDataText)
		return rr.unpackDomainName(&rr.rdataText.v, packet, startIdx)

	case QueryTypeHINFO:
		rr.rdataHINFO = new(RDataHINFO)
		endIdx := startIdx + int(rr.RDLength)
		return rr.rdataHINFO.UnmarshalBinary(packet[startIdx:endIdx])

	case QueryTypeMINFO:
		rr.rdataMINFO = new(RDataMINFO)
		return rr.unpackRDataMINFO(packet, startIdx)

	case QueryTypeMX:
		rr.rdataMX = new(RDataMX)
		return rr.unpackRDataMX(packet, startIdx)

	case QueryTypeTXT:
		rr.rdataText = new(RDataText)
		endIdx := startIdx + int(rr.RDLength)

		// The first byte of TXT is length.
		rr.rdataText.v = append(rr.rdataText.v, packet[startIdx+1:endIdx]...)

		return nil

	case QueryTypeOPT:
		rr.rdataOPT = new(RDataOPT)
		return rr.unpackRDataOPT(packet, startIdx)
	}

	return nil
}

func (rr *ResourceRecord) unpackRDataMINFO(packet []byte, startIdx int) error {
	x := startIdx
	rr.offsetIdx = 0

	err := rr.unpackDomainName(&rr.rdataMINFO.RMailBox, packet, x)
	if err != nil {
		return err
	}
	if rr.offsetIdx > 0 {
		x = rr.offsetIdx + 1
		rr.offsetIdx = 0
	} else {
		x = x + len(rr.rdataMINFO.RMailBox) + 2
	}

	err = rr.unpackDomainName(&rr.rdataMINFO.EmailBox, packet, x)
	if err != nil {
		return err
	}

	return nil
}

func (rr *ResourceRecord) unpackRDataMX(packet []byte, startIdx int) error {
	rr.rdataMX.Preference = libbytes.ReadInt16(packet, startIdx)

	rr.offsetIdx = 0
	err := rr.unpackDomainName(&rr.rdataMX.Exchange, packet, startIdx+2)

	return err
}

func (rr *ResourceRecord) unpackRDataOPT(packet []byte, x int) error {
	// Unpack extended RCODE and flags from TTL.
	rr.rdataOPT.ExtRCode = byte(rr.TTL >> 24)
	rr.rdataOPT.Version = byte(rr.TTL >> 16)

	if rr.TTL&maskOPTDO == maskOPTDO {
		rr.rdataOPT.DO = true
	}

	if rr.RDLength == 0 {
		return nil
	}

	// Unpack the RDATA
	rr.rdataOPT.Code = libbytes.ReadUint16(packet, x)
	x += 2
	rr.rdataOPT.Length = libbytes.ReadUint16(packet, x)
	x += 2
	endIdx := x + int(rr.RDLength)
	rr.rdataOPT.Data = append(rr.rdataOPT.Data, packet[x:endIdx]...)
	return nil
}

func (rr *ResourceRecord) unpackRDataSOA(packet []byte, startIdx int) error {
	x := startIdx
	rr.offsetIdx = 0

	err := rr.unpackDomainName(&rr.rdataSOA.MName, packet, x)
	if err != nil {
		return err
	}
	if rr.offsetIdx > 0 {
		x = rr.offsetIdx + 1
		rr.offsetIdx = 0
	} else {
		x = x + len(rr.rdataSOA.MName) + 2
	}

	err = rr.unpackDomainName(&rr.rdataSOA.RName, packet, x)
	if err != nil {
		return err
	}
	if rr.offsetIdx > 0 {
		x = rr.offsetIdx + 1
		rr.offsetIdx = 0
	} else {
		x = x + len(rr.rdataSOA.RName) + 2
	}

	rr.rdataSOA.Serial = libbytes.ReadUint32(packet, x)
	x += 4
	rr.rdataSOA.Refresh = libbytes.ReadInt32(packet, x)
	x += 4
	rr.rdataSOA.Retry = libbytes.ReadInt32(packet, x)
	x += 4
	rr.rdataSOA.Expire = libbytes.ReadInt32(packet, x)
	x += 4
	rr.rdataSOA.Minimum = libbytes.ReadUint32(packet, x)
	x += 4

	return nil
}
