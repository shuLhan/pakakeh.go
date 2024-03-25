// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

// ResourceRecord The answer, authority, and additional sections all share the
// same format: a variable number of resource records, where the number of
// records is specified in the corresponding count field in the header.  Each
// resource record has the following format:
type ResourceRecord struct {
	// Value hold the generic, unpacked rdata based on Type.
	Value interface{}

	// A domain name to which this resource record pertains.
	Name string

	rdata  []byte // The raw data.
	idxTTL uint16 // Position of TTL field inside packet.
	rdlen  uint16 // The length of rdata field.

	// This field specifies the meaning of the data in the rdata field.
	Type RecordType

	// The class of the data in the rdata field.
	Class RecordClass

	// A time interval (in seconds) that the resource record may be cached
	// before it should be discarded.
	// Zero values are interpreted to mean that the RR can only be used
	// for the transaction in progress, and should not be cached.
	TTL uint32
}

// String return the text representation of ResourceRecord for human.
func (rr *ResourceRecord) String() string {
	return fmt.Sprintf("{Name:%s Type:%d Class:%d TTL:%d Value:%v}",
		rr.Name, rr.Type, rr.Class, rr.TTL, rr.Value)
}

// initAndValidate initialize and validate the resource record data.
// It will return an error if one of the required fields is empty or if its
// type is not match with its value.
func (rr *ResourceRecord) initAndValidate() (err error) {
	var (
		logp = "initAndValidate"

		mx    *RDataMX
		soa   *RDataSOA
		srv   *RDataSRV
		ip    net.IP
		ipv4  net.IP
		ipv6  net.IP
		v     string
		rtype string
		ok    bool
	)

	if len(rr.Name) == 0 {
		return fmt.Errorf("%s: empty Name", logp)
	}
	if rr.Class == 0 {
		rr.Class = RecordClassIN
	}
	if rr.TTL == 0 {
		rr.TTL = defaultTTL
	}

	rtype, ok = RecordTypeNames[rr.Type]
	if !ok {
		return fmt.Errorf("%s: unknown type %d", logp, rr.Type)
	}
	switch rr.Type {
	case RecordTypeA:
		v, ok = rr.Value.(string)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("%s: invalid or empty %s: %q", logp, rtype, v)
		}
		ipv4 = ip.To4()
		if ipv4 == nil {
			return fmt.Errorf("%s: invalid or empty %s: %q", logp, rtype, v)
		}

	case RecordTypeNS, RecordTypeCNAME, RecordTypeMB, RecordTypeMG,
		RecordTypeMR, RecordTypeNULL, RecordTypePTR:

		v, ok = rr.Value.(string)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}

		if !libnet.IsHostnameValid([]byte(v), true) {
			return fmt.Errorf("%s: invalid or empty %s: %q", logp, rtype, v)
		}

	case RecordTypeSOA:
		soa, ok = rr.Value.(*RDataSOA)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
		if !libnet.IsHostnameValid([]byte(soa.MName), true) {
			return fmt.Errorf("%s: invalid or empty %s MName: %q", logp, rtype, soa.MName)
		}
		if !libnet.IsHostnameValid([]byte(soa.RName), true) {
			return fmt.Errorf("%s: invalid or empty %s RName: %q", logp, rtype, soa.RName)
		}
	case RecordTypeWKS:
		_, ok = rr.Value.(*RDataWKS)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}

	case RecordTypeHINFO:
		_, ok = rr.Value.(*RDataHINFO)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
	case RecordTypeMINFO:
		_, ok = rr.Value.(*RDataMINFO)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
	case RecordTypeMX:
		mx, ok = rr.Value.(*RDataMX)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
		err = mx.initAndValidate()
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	case RecordTypeTXT:
		v, ok = rr.Value.(string)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
		if len(v) == 0 {
			return fmt.Errorf("%s: empty %s value", logp, rtype)
		}
	case RecordTypeSRV:
		srv, ok = rr.Value.(*RDataSRV)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
		err = srv.initAndValidate()
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	case RecordTypeAAAA:
		v, ok = rr.Value.(string)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
		ip = net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("%s: invalid or empty %s value: %q", logp, rtype, v)
		}
		ipv6 = ip.To16()
		if ipv6 == nil {
			return fmt.Errorf("%s: invalid or empty %s value: %q", logp, rtype, v)
		}
	case RecordTypeOPT:
		_, ok = rr.Value.(*RDataOPT)
		if !ok {
			return fmt.Errorf("%s: expecting %s got %T", logp, rtype, rr.Value)
		}
	}
	return nil
}

// unpack the resource record from packet start from index startIdx.
func (rr *ResourceRecord) unpack(packet []byte, startIdx uint) (x uint, err error) {
	var (
		logp      = "ResourceRecord.unpack"
		lenPacket = uint(len(packet))

		lenXRdata uint
	)

	x = startIdx

	rr.Name, x, err = unpackDomainName(packet, x)
	if err != nil {
		return x, fmt.Errorf("%s: %w", logp, err)
	}

	rr.Type = RecordType(libbytes.ReadUint16(packet, x))
	x += 2
	rr.Class = RecordClass(libbytes.ReadUint16(packet, x))
	x += 2
	rr.idxTTL = uint16(x)
	rr.TTL = libbytes.ReadUint32(packet, x)
	x += 4
	rr.rdlen = libbytes.ReadUint16(packet, x)
	x += 2

	lenXRdata = x + uint(rr.rdlen)
	if lenPacket < lenXRdata {
		return x, fmt.Errorf("%s: %s %d: packet length %d smaller than index+rdata %d+%d (%d)",
			logp, rr.Name, rr.Type, lenPacket, x, rr.rdlen, lenXRdata)
	}

	rr.rdata = append(rr.rdata, packet[x:lenXRdata]...)

	err = rr.unpackRData(packet, x)
	if err != nil {
		return x, fmt.Errorf("%s: %w", logp, err)
	}

	x += uint(rr.rdlen)

	return x, nil
}

// unpackDomainName unpack domain name from packet from index start.
// It will return the domain name and the last index where domain name end.
func unpackDomainName(packet []byte, start uint) (name string, end uint, err error) {
	var (
		logp = `unpackDomainName`
		x    = int(start)

		out      strings.Builder
		count, y byte
		isJump   bool
	)

	end = start

	for x < len(packet) {
		count = packet[x]
		if count == 0 {
			if !isJump {
				end++
			}
			break
		}
		if (packet[x] & maskPointer) == maskPointer {
			var offset = uint16(packet[x]&maskOffset)<<8 | uint16(packet[x+1])
			// Jump to index defined by offset.
			x = int(offset)
			if !isJump {
				end += 2
				isJump = true
			}
			continue
		}
		if count > maxLabelSize {
			return ``, end, fmt.Errorf(`%s: at %d: %w`, logp, x, ErrLabelSizeLimit)
		}

		x++
		if !isJump {
			end++
		}
		if x+int(count) >= len(packet) {
			// It should not goes here, invalid packet.
			return ``, end, fmt.Errorf(`%s: label size %d greater than packet length %d`, logp, count, len(packet))
		}

		if out.Len() > 0 {
			out.WriteByte('.')
		}
		for y = 0; y < count; y++ {
			if packet[x] >= 'A' && packet[x] <= 'Z' {
				out.WriteByte(packet[x] + 32)
			} else {
				out.WriteByte(packet[x])
			}
			x++
		}
		if !isJump {
			end += uint(count)
		}
	}
	return out.String(), end, nil
}

func (rr *ResourceRecord) unpackRData(packet []byte, startIdx uint) (err error) {
	var (
		rrWKS   *RDataWKS
		rrHInfo *RDataHINFO
		endIdx  uint
	)

	switch rr.Type {
	case RecordTypeA:
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
	case RecordTypeNS:
		rr.Value, _, err = unpackDomainName(packet, startIdx)
		return err

	// MD is obsolete.  See the definition of MX and [RFC-974] for details of
	// the new scheme.  The recommended policy for dealing with MD RRs found in
	// a zone file is to reject them, or to convert them to MX RRs with a
	// preference of 0.
	case RecordTypeMD:

	// MF is obsolete.  See the definition of MX and [RFC-974] for details
	// ofw the new scheme.  The recommended policy for dealing with MD RRs
	// found in a zone file is to reject them, or to convert them to MX
	// RRs with a preference of 10.
	case RecordTypeMF:

	// CNAME RRs cause no additional section processing, but name servers
	// may choose to restart the query at the canonical name in certain
	// cases.  See the description of name server logic in [RFC-1034] for
	// details.
	case RecordTypeCNAME:
		rr.Value, _, err = unpackDomainName(packet, startIdx)
		return err

	case RecordTypeSOA:
		return rr.unpackSOA(packet, startIdx)

	case RecordTypeMB:
		rr.Value, _, err = unpackDomainName(packet, startIdx)
		return err

	case RecordTypeMG:
		rr.Value, _, err = unpackDomainName(packet, startIdx)
		return err

	case RecordTypeMR:
		rr.Value, _, err = unpackDomainName(packet, startIdx)
		return err

	// NULL records cause no additional section processing.
	// NULLs are used as placeholders in some experimental extensions of
	// the DNS.
	case RecordTypeNULL:
		endIdx = startIdx + uint(rr.rdlen)
		rr.Value = string(packet[startIdx : startIdx+endIdx])
		return nil

	case RecordTypeWKS:
		rrWKS = new(RDataWKS)
		rr.Value = rrWKS
		endIdx = startIdx + uint(rr.rdlen)
		return rrWKS.unpack(packet[startIdx:endIdx])

	case RecordTypePTR:
		rr.Value, _, err = unpackDomainName(packet, startIdx)
		return err

	case RecordTypeHINFO:
		rrHInfo = new(RDataHINFO)
		rr.Value = rrHInfo
		endIdx = startIdx + uint(rr.rdlen)
		return rrHInfo.unpack(packet[startIdx:endIdx])

	case RecordTypeMINFO:
		return rr.unpackMInfo(packet, startIdx)

	case RecordTypeMX:
		return rr.unpackMX(packet, startIdx)

	case RecordTypeTXT:
		endIdx = startIdx + uint(rr.rdlen)

		// The first byte of TXT is length.
		rr.Value = string(packet[startIdx+1 : endIdx])

		return nil

	case RecordTypeAAAA:
		return rr.unpackAAAA()

	case RecordTypeSRV:
		return rr.unpackSRV(packet, startIdx)

	case RecordTypeOPT:
		return rr.unpackOPT(packet, startIdx)

	case RecordTypeSVCB:
		return rr.unpackSVCB(packet, startIdx)

	case RecordTypeHTTPS:
		return rr.unpackHTTPS(packet, startIdx)

	default:
		log.Printf("= Unknown query type: %d\n", rr.Type)
	}

	return nil
}

func (rr *ResourceRecord) unpackA() error {
	if rr.rdlen != rdataIPv4Size || len(rr.rdata) != rdataIPv4Size {
		return ErrIPv4Length
	}

	var ip = net.IP(rr.rdata)

	rr.Value = ip.String()

	return nil
}

func (rr *ResourceRecord) unpackAAAA() error {
	if rr.rdlen != rdataIPv6Size || len(rr.rdata) != rdataIPv6Size {
		return ErrIPv6Length
	}

	var ip = net.IP(rr.rdata)
	rr.Value = ip.String()

	return nil
}

func (rr *ResourceRecord) unpackMInfo(packet []byte, startIdx uint) (err error) {
	var (
		logp    = "unpackMInfo"
		rrMInfo = &RDataMINFO{}
		x       = startIdx
	)

	rr.Value = rrMInfo

	rrMInfo.RMailBox, x, err = unpackDomainName(packet, x)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	rrMInfo.EmailBox, _, err = unpackDomainName(packet, x)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

func (rr *ResourceRecord) unpackMX(packet []byte, startIdx uint) (err error) {
	var (
		rrMX = &RDataMX{}
	)

	rr.Value = rrMX

	rrMX.Preference = libbytes.ReadInt16(packet, startIdx)

	rrMX.Exchange, _, err = unpackDomainName(packet, startIdx+2)

	return err
}

func (rr *ResourceRecord) unpackSRV(packet []byte, x uint) (err error) {
	var (
		rrSRV = &RDataSRV{}

		start int
		y     int
	)

	rr.Value = rrSRV

	// Unpack service, proto, and name from RR.Name
	for ; y < len(rr.Name); y++ {
		if rr.Name[y] == '.' {
			rrSRV.Service = rr.Name[start:y]
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

	rrSRV.Target, _, err = unpackDomainName(packet, x)

	return
}

func (rr *ResourceRecord) unpackOPT(packet []byte, x uint) error {
	var (
		rrOPT = &RDataOPT{}

		endIdx uint
	)

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
	endIdx = x + uint(rr.rdlen)
	if int(endIdx) >= len(packet) {
		return errors.New("unpackOPT: data length is out of range")
	}
	rrOPT.Data = append(rrOPT.Data, packet[x:endIdx]...)
	return nil
}

func (rr *ResourceRecord) unpackSVCB(packet []byte, x uint) (err error) {
	var (
		logp = `unpackSVCB`
		svcb = &RDataSVCB{
			Params: map[int][]string{},
		}
	)

	packet = packet[x:]

	err = svcb.unpack(packet)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	rr.Value = svcb

	return nil
}

func (rr *ResourceRecord) unpackHTTPS(packet []byte, x uint) (err error) {
	var (
		logp  = `unpackHTTPS`
		https = &RDataHTTPS{
			RDataSVCB: RDataSVCB{
				Params: map[int][]string{},
			},
		}
	)

	packet = packet[x:]

	err = https.RDataSVCB.unpack(packet)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	rr.Value = https

	return nil
}

func (rr *ResourceRecord) unpackSOA(packet []byte, startIdx uint) (err error) {
	var (
		logp  = "unpackSOA"
		rrSOA = &RDataSOA{}
		x     = startIdx
		end   uint
	)

	rr.Value = rrSOA

	rrSOA.MName, end, err = unpackDomainName(packet, x)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if end > 0 {
		x = end
	} else {
		x += uint(len(rrSOA.MName) + 2)
	}

	rrSOA.RName, end, err = unpackDomainName(packet, x)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if end > 0 {
		x = end
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
