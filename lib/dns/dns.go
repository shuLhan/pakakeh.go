// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dns implement DNS client and server, as defined by RFC 1035.
package dns

import (
	"errors"
	"time"
)

const (
	// Port define default DNS remote or listen port.
	Port = 53

	maskPointer byte  = 0xC0
	maskOffset  byte  = 0x3F
	maskOPTDO   int32 = 0x00008000

	maxLabelSize     = 63
	maxUDPPacketSize = 512
	rdataAddrSize    = 4
	// sectionHeaderSize define the size of section header in DNS message.
	sectionHeaderSize = 12
)

var (
	// NameServers define default name servers.
	NameServers = []string{
		"127.0.0.1:53",
	}

	// clientTimeout define read and write timeout on client request.
	clientTimeout = 3 * time.Second
	debugLevel    = 0
)

//
// List of error messages.
//
var (
	ErrNewConnection   = errors.New("Lookup: can't create new connection")
	ErrLabelSizeLimit  = errors.New("Labels should be 63 octet or less")
	ErrRDataAddrLength = errors.New("Invalid length of RData A format")
)

type OpCode byte

const (
	OpCodeQuery  OpCode = iota // a standard query (QUERY)
	OpCodeIQuery               // an inverse query (IQUERY)
	OpCodeStatus               // a server status request (STATUS)
)

// QueryType define type of query in section question and in resource records.
type QueryType uint16

// List of query types.
const (
	QueryTypeZERO  QueryType = iota // Empty query type.
	QueryTypeA                      // A host address
	QueryTypeNS                     // An authoritative name server
	QueryTypeMD                     // A mail destination (Obsolete - use MX)
	QueryTypeMF                     // A mail forwarder (Obsolete - use MX)
	QueryTypeCNAME                  // The canonical name for an alias
	QueryTypeSOA                    // Marks the start of a zone of authority
	QueryTypeMB                     // A mailbox domain name (EXPERIMENTAL)
	QueryTypeMG                     // A mail group member (EXPERIMENTAL)
	QueryTypeMR                     // A mail rename domain name (EXPERIMENTAL)
	QueryTypeNULL                   // A null RR (EXPERIMENTAL)
	QueryTypeWKS                    // A well known service description
	QueryTypePTR                    // A domain name pointer
	QueryTypeHINFO                  // Host information
	QueryTypeMINFO                  // Mailbox or mail list information
	QueryTypeMX                     // Mail exchange
	QueryTypeTXT                    // (16) Text strings
	QueryTypeOPT   QueryType = 41   // An OPT pseudo-RR (sometimes called a meta-RR)
	QueryTypeAXFR  QueryType = 252  // A request for a transfer of an entire zone
	QueryTypeMAILB QueryType = 253  // A request for mailbox-related records (MB, MG or MR)
	QueryTypeMAILA QueryType = 254  // A request for mail agent RRs (Obsolete - see MX)
	QueryTypeALL   QueryType = 255  // A request for all records
)

// QueryClass define a two octet code that specifies the class of the query.
type QueryClass uint16

const (
	QueryClassZERO QueryClass = iota // Empty query class.
	QueryClassIN                     // The Internet
	QueryClassCS                     // The CSNET class (Obsolete - used only for examples in some obsolete RFCs)
	QueryClassCH                     // The CHAOS class
	QueryClassHS                     // Hesiod [Dyer 87]
	QueryClassANY  QueryClass = 255  // Any class
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
