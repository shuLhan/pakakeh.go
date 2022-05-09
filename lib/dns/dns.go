// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dns implement DNS client and server.
//
// This library implemented in reference to,
//
//   - RFC1034 DOMAIN NAMES - CONCEPTS AND FACILITIES
//   - RFC1035 DOMAIN NAMES - IMPLEMENTATION AND SPECIFICATION
//   - RFC1886 DNS Extensions to support IP version 6.
//   - RFC2782 A DNS RR for specifying the location of services (DNS SRV)
//   - RFC6891 Extension Mechanisms for DNS (EDNS(0))
//   - RFC8484 DNS Queries over HTTPS (DoH)
package dns

import (
	"errors"
	"time"
)

const (
	// DefaultPort define default DNS remote or listen port for UDP and
	// TCP connection.
	DefaultPort uint16 = 53

	// DefaultTLSPort define default remote and listen port for DNS over
	// TLS.
	DefaultTLSPort uint16 = 853

	// DefaultHTTPPort define default port for DNS over HTTPS.
	DefaultHTTPPort        uint16        = 443
	defaultHTTPIdleTimeout time.Duration = 120 * time.Second
)

const (
	maskPointer byte   = 0xC0
	maskOffset  byte   = 0x3F
	maskOPTDO   uint32 = 0x00008000

	maxLabelSize     = 63
	maxUdpPacketSize = 1232
	maxTcpPacketSize = 4096
	rdataIPv4Size    = 4
	rdataIPv6Size    = 16
	// sectionHeaderSize define the size of section header in DNS message.
	sectionHeaderSize = 12

	dohHeaderKeyAccept      = "accept"
	dohHeaderKeyContentType = "content-type"
	dohHeaderValDNSMessage  = "application/dns-message"
)

// List of error messages.
var (
	ErrNewConnection  = errors.New("lookup: can't create new connection")
	ErrLabelSizeLimit = errors.New("labels should be 63 octet or less")
	ErrInvalidAddress = errors.New("invalid address")
	ErrIPv4Length     = errors.New("invalid length of A RDATA format")
	ErrIPv6Length     = errors.New("invalid length of AAAA RDATA format")
)

var (
	// clientTimeout define read and write timeout on client request.
	clientTimeout = 6 * time.Second
)

type connType byte

// List of known connection type.
const (
	connTypeUDP connType = 1 << iota
	connTypeTCP
	connTypeDoH
	connTypeDoT
)

// connTypeNames contains a mapping between connection type and its name.
var connTypeNames = map[connType]string{
	connTypeUDP: "UDP",
	connTypeTCP: "TCP",
	connTypeDoH: "DoH",
	connTypeDoT: "DoT",
}

// OpCode define a custom type for DNS header operation code.
type OpCode byte

// List of valid operation code.
const (
	OpCodeQuery  OpCode = iota // A standard query (QUERY)
	OpCodeIQuery               // An inverse query (IQUERY), obsolete by RFC3425
	OpCodeStatus               // A server status request (STATUS)
)

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

// rcodeNames contains mapping of response code with their human readable
// names.
var rcodeNames = map[ResponseCode]string{
	RCodeOK:             "OK",
	RCodeErrFormat:      "ERR_FORMAT",
	RCodeErrServer:      "ERR_SERVER",
	RCodeErrName:        "ERR_NAME",
	RCodeNotImplemented: "ERR_NOT_IMPLEMENTED",
	RCodeRefused:        "ERR_REFUSED",
}
