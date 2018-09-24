// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dns implement DNS client and server.
//
// This library implemented in reference to,
//
//	- RFC1034 DOMAIN NAMES - CONCEPTS AND FACILITIES
//	- RFC1035 DOMAIN NAMES - IMPLEMENTATION AND SPECIFICATION
//	- RFC1886 DNS Extensions to support IP version 6.
//	- RFC2782 A DNS RR for specifying the location of services (DNS SRV)
//	- RFC6891 Extension Mechanisms for DNS (EDNS(0))
//
package dns

import (
	"errors"
	"net"
	"time"

	libnet "github.com/shuLhan/share/lib/net"
)

const (
	// DefaultPort define default DNS remote or listen port for UDP and
	// TCP connection.
	DefaultPort uint16 = 53

	// DefaultDoHPort define default port for DoH.
	DefaultDoHPort        uint16        = 443
	defaultDoHIdleTimeout time.Duration = 120 * time.Second
)

const (
	maskPointer byte   = 0xC0
	maskOffset  byte   = 0x3F
	maskOPTDO   uint32 = 0x00008000

	maxLabelSize     = 63
	maxUDPPacketSize = 4096
	rdataIPv4Size    = 4
	rdataIPv6Size    = 16
	// sectionHeaderSize define the size of section header in DNS message.
	sectionHeaderSize = 12

	dohHeaderKeyAccept      = "accept"
	dohHeaderKeyContentType = "content-type"
	dohHeaderValDNSMessage  = "application/dns-message"
)

//
// List of error messages.
//
var (
	ErrNewConnection  = errors.New("Lookup: can't create new connection")
	ErrLabelSizeLimit = errors.New("Labels should be 63 octet or less")
	ErrInvalidAddress = errors.New("Invalid address")
	ErrIPv4Length     = errors.New("Invalid length of A RDATA format")
	ErrIPv6Length     = errors.New("Invalid length of AAAA RDATA format")
)

var (
	// clientTimeout define read and write timeout on client request.
	clientTimeout = 6 * time.Second
	debugLevel    = 0
)

type OpCode byte

const (
	OpCodeQuery  OpCode = iota // a standard query (QUERY)
	OpCodeIQuery               // an inverse query (IQUERY), obsolete by RFC3425
	OpCodeStatus               // a server status request (STATUS)
)

// List of code for known DNS query types.
const (
	QueryTypeZERO  uint16 = iota // Empty query type.
	QueryTypeA                   // A host address
	QueryTypeNS                  // An authoritative name server
	QueryTypeMD                  // A mail destination (Obsolete - use MX)
	QueryTypeMF                  // A mail forwarder (Obsolete - use MX)
	QueryTypeCNAME               // The canonical name for an alias
	QueryTypeSOA                 // Marks the start of a zone of authority
	QueryTypeMB                  // A mailbox domain name (EXPERIMENTAL)
	QueryTypeMG                  // A mail group member (EXPERIMENTAL)
	QueryTypeMR                  // A mail rename domain name (EXPERIMENTAL)
	QueryTypeNULL                // A null RR (EXPERIMENTAL)
	QueryTypeWKS                 // A well known service description
	QueryTypePTR                 // A domain name pointer
	QueryTypeHINFO               // Host information
	QueryTypeMINFO               // Mailbox or mail list information
	QueryTypeMX                  // Mail exchange
	QueryTypeTXT                 // (16) Text strings
	QueryTypeAAAA  uint16 = 28   // IPv6 address
	QueryTypeSRV   uint16 = 33   // A SRV RR for locating service.
	QueryTypeOPT   uint16 = 41   // An OPT pseudo-RR (sometimes called a meta-RR)
	QueryTypeAXFR  uint16 = 252  // A request for a transfer of an entire zone
	QueryTypeMAILB uint16 = 253  // A request for mailbox-related records (MB, MG or MR)
	QueryTypeMAILA uint16 = 254  // A request for mail agent RRs (Obsolete - see MX)
	QueryTypeALL   uint16 = 255  // A request for all records
)

//
// QueryTypes contains a mapping between string representation of DNS query
// type with their decimal value.
//
var QueryTypes = map[string]uint16{
	"A":     QueryTypeA,
	"NS":    QueryTypeNS,
	"CNAME": QueryTypeCNAME,
	"SOA":   QueryTypeSOA,
	"MB":    QueryTypeMB,
	"MG":    QueryTypeMG,
	"MR":    QueryTypeMR,
	"NULL":  QueryTypeNULL,
	"WKS":   QueryTypeWKS,
	"PTR":   QueryTypePTR,
	"HINFO": QueryTypeHINFO,
	"MINFO": QueryTypeMINFO,
	"MX":    QueryTypeMX,
	"TXT":   QueryTypeTXT,
	"AAAA":  QueryTypeAAAA,
	"SRV":   QueryTypeSRV,
	"OPT":   QueryTypeOPT,
}

// List of code known DNS query class.
const (
	QueryClassZERO uint16 = iota // Empty query class.
	QueryClassIN                 // The Internet
	QueryClassCS                 // The CSNET class (Obsolete - used only for examples in some obsolete RFCs)
	QueryClassCH                 // The CHAOS class
	QueryClassHS                 // Hesiod [Dyer 87]
	QueryClassANY  uint16 = 255  // Any class
)

//
// QueryClasses contains a mapping between string representation of DNS query
// class with their decimal value.
//
var QueryClasses = map[string]uint16{
	"IN": QueryClassIN,
	"CH": QueryClassCH,
	"HS": QueryClassHS,
}

// ResponseCode define response code in message header.
type ResponseCode byte

// List of response codes.
const (
	RCodeOK ResponseCode = iota //  No error condition

	// Format error - The name server was unable to interpret the query.
	RCodeErrFormat

	// Server failure - The name server was unable to process this query
	// due to a problem with the name server.
	RCodeErrServer

	// Name Error - Meaningful only for responses from an authoritative
	// name server, this code signifies that the domain name referenced in
	// the query does not exist.
	RCodeErrName

	// Not Implemented - The name server does not support the requested
	// kind of query.
	RCodeNotImplemented

	// Refused - The name server refuses to perform the specified
	// operation for policy reasons.  For example, a name server may not
	// wish to provide the information to the particular requester, or a
	// name server may not wish to perform a particular operation (e.g.,
	// zone transfer) for particular data.
	RCodeRefused
)

//
// ParseNameServers parse list of nameserver into UDP addresses.
// If one of nameserver is invalid it will stop parsing and return only valid
// nameserver addresses with error.
//
func ParseNameServers(nameservers []string) ([]*net.UDPAddr, error) {
	udpAddrs := make([]*net.UDPAddr, 0)

	for _, ns := range nameservers {
		addr, err := libnet.ParseUDPAddr(ns, DefaultPort)
		if err != nil {
			return udpAddrs, err
		}
		udpAddrs = append(udpAddrs, addr)
	}

	return udpAddrs, nil
}
