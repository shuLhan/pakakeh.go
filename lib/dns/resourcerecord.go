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
	"strings"

	libbytes "github.com/shuLhan/share/lib/bytes"
	libnet "github.com/shuLhan/share/lib/net"
)

//
// ResourceRecord The answer, authority, and additional sections all share the
// same format: a variable number of resource records, where the number of
// records is specified in the corresponding count field in the header.  Each
// resource record has the following format:
//
type ResourceRecord struct {
	// A domain name to which this resource record pertains.
	Name string

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

	// Value hold the generic value for all record types.
	Value interface{}

	// An unsigned 16 bit integer that specifies the length in octets of
	// the RDATA field.
	rdlen uint16

	// A variable length string of octets that describes the resource.
	// The format of this information varies according to the TYPE and
	// CLASS of the resource record.  For example, if the TYPE is A
	// and the CLASS is IN, the RDATA field is a 4 octet ARPA Internet
	// address.
	rdata []byte

	off    uint
	offTTL uint
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
// initAndValidate initialize and validate the resource record data.
// It will return an error if one of the required fields is empty or if its
// type is not match with its value.
//
func (rr *ResourceRecord) initAndValidate() (err error) {
	if len(rr.Name) == 0 {
		return errors.New("empty RR name")
	}
	if rr.Class == 0 {
		rr.Class = QueryClassIN
	}
	if rr.TTL == 0 {
		rr.TTL = defaultTTL
	}

	switch rr.Type {
	case QueryTypeA:
		v, ok := rr.Value.(string)
		if !ok {
			return fmt.Errorf("RR A: expecting Value as string got %T", rr.Value)
		}
		s := string(v)
		ip := net.ParseIP(s)
		if ip == nil {
			return fmt.Errorf("RR A: invalid or empty IP address %q", s)
		}
		ipv4 := ip.To4()
		if ipv4 == nil {
			return fmt.Errorf("RR A: invalid or empty IPv4 address %q", s)
		}

	case QueryTypeNS, QueryTypeCNAME, QueryTypeMB, QueryTypeMG,
		QueryTypeMR, QueryTypeNULL, QueryTypePTR:

		v, ok := rr.Value.(string)
		if !ok {
			return fmt.Errorf("RR %s: expecting Value as string got %T",
				QueryTypeNames[rr.Type], rr.Value)
		}

		if !libnet.IsHostnameValid([]byte(v), true) {
			return fmt.Errorf("RR %s: invalid or empty value: %q",
				QueryTypeNames[rr.Type], v)
		}

	case QueryTypeSOA:
		soa, ok := rr.Value.(*RDataSOA)
		if !ok {
			return fmt.Errorf("RR SOA: expecting RDataSOA got %T", rr.Value)
		}
		if !libnet.IsHostnameValid([]byte(soa.MName), true) {
			return fmt.Errorf("RR SOA: invalid or empty MName: %q", soa.MName)
		}
		if !libnet.IsHostnameValid([]byte(soa.RName), true) {
			return fmt.Errorf("RR SOA: invalid or empty RName: %q", soa.RName)
		}
	case QueryTypeWKS:
		_, ok := rr.Value.(*RDataWKS)
		if !ok {
			return fmt.Errorf("RR WKS: expecting WKS got %T", rr.Value)
		}

	case QueryTypeHINFO:
		_, ok := rr.Value.(*RDataHINFO)
		if !ok {
			return fmt.Errorf("RR HINFO: expecting HINFO got %T", rr.Value)
		}
	case QueryTypeMINFO:
		_, ok := rr.Value.(*RDataMINFO)
		if !ok {
			return fmt.Errorf("RR MINFO: expecting MINFO got %T", rr.Value)
		}
	case QueryTypeMX:
		mx, ok := rr.Value.(*RDataMX)
		if !ok {
			return fmt.Errorf("RR MX: expecting MX got %T", rr.Value)
		}
		err = mx.initAndValidate()
		if err != nil {
			return err
		}
	case QueryTypeTXT:
		txt, ok := rr.Value.(string)
		if !ok {
			return fmt.Errorf("RR TXT: expecting string got %T", rr.Value)
		}
		if len(txt) == 0 {
			return errors.New("empty RR TXT value")
		}
	case QueryTypeSRV:
		srv, ok := rr.Value.(*RDataSRV)
		if !ok {
			return fmt.Errorf("RR SRV: expecting SRV got %T", rr.Value)
		}
		err = srv.initAndValidate()
		if err != nil {
			return err
		}
	case QueryTypeAAAA:
		v, ok := rr.Value.(string)
		if !ok {
			return fmt.Errorf("RR AAAA: expecting AAAA got %T", rr.Value)
		}
		s := string(v)
		ip := net.ParseIP(s)
		if ip == nil {
			return fmt.Errorf("RR AAAA: invalid or empty IPv6 address: %q", s)
		}
		ipv6 := ip.To16()
		if ipv6 == nil {
			return fmt.Errorf("RR AAAA: invalid or empty IPv6 address: %q", s)
		}
	case QueryTypeOPT:
		_, ok := rr.Value.(*RDataOPT)
		if !ok {
			return fmt.Errorf("RR OPT: expecting OPT got %T", rr.Value)
		}
	default:
		return fmt.Errorf("unknown RR type %d", rr.Type)
	}
	return nil
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
	string, error,
) {
	var out strings.Builder

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
			return "", ErrLabelSizeLimit
		}
		if out.Len() > 0 {
			out.WriteByte('.')
		}

		x++
		for y := byte(0); y < count; y++ {
			if x >= len(packet) {
				break
			}
			if packet[x] >= 'A' && packet[x] <= 'Z' {
				packet[x] += 32
			}
			out.WriteByte(packet[x])
			x++
		}
	}
	return out.String(), nil
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
		rr.Value, err = rr.unpackDomainName(packet, startIdx)
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
		rr.Value, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeSOA:
		return rr.unpackSOA(packet, startIdx)

	case QueryTypeMB:
		rr.Value, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeMG:
		rr.Value, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeMR:
		rr.Value, err = rr.unpackDomainName(packet, startIdx)
		return err

	// NULL records cause no additional section processing.
	// NULLs are used as placeholders in some experimental extensions of
	// the DNS.
	case QueryTypeNULL:
		endIdx := startIdx + uint(rr.rdlen)
		rr.Value = string(packet[startIdx : startIdx+endIdx])
		return nil

	case QueryTypeWKS:
		rrWKS := new(RDataWKS)
		rr.Value = rrWKS
		endIdx := startIdx + uint(rr.rdlen)
		return rrWKS.unpack(packet[startIdx:endIdx])

	case QueryTypePTR:
		rr.Value, err = rr.unpackDomainName(packet, startIdx)
		return err

	case QueryTypeHINFO:
		rrHInfo := new(RDataHINFO)
		rr.Value = rrHInfo
		endIdx := startIdx + uint(rr.rdlen)
		return rrHInfo.unpack(packet[startIdx:endIdx])

	case QueryTypeMINFO:
		return rr.unpackMInfo(packet, startIdx)

	case QueryTypeMX:
		return rr.unpackMX(packet, startIdx)

	case QueryTypeTXT:
		endIdx := startIdx + uint(rr.rdlen)

		// The first byte of TXT is length.
		rr.Value = string(packet[startIdx+1 : endIdx])

		return nil

	case QueryTypeAAAA:
		return rr.unpackAAAA()

	case QueryTypeSRV:
		return rr.unpackSRV(packet, startIdx)

	case QueryTypeOPT:
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
	rr.Value = ip.String()

	return nil
}

func (rr *ResourceRecord) unpackAAAA() error {
	if rr.rdlen != rdataIPv6Size || len(rr.rdata) != rdataIPv6Size {
		return ErrIPv6Length
	}

	ip := net.IP(rr.rdata)
	rr.Value = ip.String()

	return nil
}

func (rr *ResourceRecord) unpackMInfo(packet []byte, startIdx uint) (err error) {
	x := startIdx
	rr.off = 0

	rrMInfo := &RDataMINFO{}
	rr.Value = rrMInfo

	rrMInfo.RMailBox, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x += uint(len(rrMInfo.RMailBox) + 2)
	}

	rrMInfo.EmailBox, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}

	return nil
}

func (rr *ResourceRecord) unpackMX(packet []byte, startIdx uint) (err error) {
	rrMX := &RDataMX{}
	rr.Value = rrMX

	rrMX.Preference = libbytes.ReadInt16(packet, startIdx)

	rr.off = 0
	rrMX.Exchange, err = rr.unpackDomainName(packet, startIdx+2)

	return err
}

func (rr *ResourceRecord) unpackSRV(packet []byte, x uint) (err error) {
	rrSRV := &RDataSRV{}
	rr.Value = rrSRV

	// Unpack service, proto, and name from RR.Name
	start := 0
	y := 0
	for ; y < len(rr.Name); y++ {
		if rr.Name[y] == '.' {
			rrSRV.Service = string(rr.Name[start:y])
			break
		}
	}
	y++
	start = y
	for ; y < len(rr.Name); y++ {
		if rr.Name[y] == '.' {
			rrSRV.Proto = rr.Name[start:y]
			break
		}
	}
	y++
	rrSRV.Name = rr.Name[y:]

	// Unpack RDATA
	rrSRV.Priority = libbytes.ReadUint16(packet, x)
	x += 2
	rrSRV.Weight = libbytes.ReadUint16(packet, x)
	x += 2
	rrSRV.Port = libbytes.ReadUint16(packet, x)
	x += 2

	rrSRV.Target, err = rr.unpackDomainName(packet, x)

	return
}

func (rr *ResourceRecord) unpackOPT(packet []byte, x uint) error {
	rrOPT := &RDataOPT{}
	rr.Value = rrOPT

	// Unpack extended RCODE and flags from TTL.
	rrOPT.ExtRCode = byte(rr.TTL >> 24)
	rrOPT.Version = byte(rr.TTL >> 16)

	if rr.TTL&maskOPTDO == maskOPTDO {
		rrOPT.DO = true
	}

	if rr.rdlen == 0 {
		return nil
	}

	// Unpack the RDATA
	rrOPT.Code = libbytes.ReadUint16(packet, x)
	x += 2
	rrOPT.Length = libbytes.ReadUint16(packet, x)
	x += 2
	endIdx := x + uint(rr.rdlen)
	if int(endIdx) >= len(packet) {
		return errors.New("RR OPT length is out of range")
	}
	rrOPT.Data = append(rrOPT.Data, packet[x:endIdx]...)
	return nil
}

func (rr *ResourceRecord) unpackSOA(packet []byte, startIdx uint) (err error) {
	rrSOA := &RDataSOA{}
	rr.Value = rrSOA

	x := startIdx
	rr.off = 0

	rrSOA.MName, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x += uint(len(rrSOA.MName) + 2)
	}

	rrSOA.RName, err = rr.unpackDomainName(packet, x)
	if err != nil {
		return err
	}
	if rr.off > 0 {
		x = rr.off + 1
		rr.off = 0
	} else {
		x += uint(len(rrSOA.RName) + 2)
	}

	rrSOA.Serial = libbytes.ReadUint32(packet, x)
	x += 4
	rrSOA.Refresh = libbytes.ReadInt32(packet, x)
	x += 4
	rrSOA.Retry = libbytes.ReadInt32(packet, x)
	x += 4
	rrSOA.Expire = libbytes.ReadInt32(packet, x)
	x += 4
	rrSOA.Minimum = libbytes.ReadUint32(packet, x)

	return nil
}
