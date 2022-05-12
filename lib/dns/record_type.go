// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import "net"

// RecordType A two octet code which specifies the type of the record.
type RecordType uint16

// List of code for known DNS record types, ordered by value.
const (
	RecordTypeZERO  RecordType = iota // Empty record type.
	RecordTypeA                       //  1 - A host address
	RecordTypeNS                      //  2 - An authoritative name server
	RecordTypeMD                      //  3 - A mail destination (Obsolete - use MX)
	RecordTypeMF                      //  4 - A mail forwarder (Obsolete - use MX)
	RecordTypeCNAME                   //  5 - The canonical name for an alias
	RecordTypeSOA                     //  6 - Marks the start of a zone of authority
	RecordTypeMB                      //  7 - A mailbox domain name (EXPERIMENTAL)
	RecordTypeMG                      //  8 - A mail group member (EXPERIMENTAL)
	RecordTypeMR                      //  9 - A mail rename domain name (EXPERIMENTAL)
	RecordTypeNULL                    // 10 - A null RR (EXPERIMENTAL)
	RecordTypeWKS                     // 11 - A well known service description
	RecordTypePTR                     // 12 - A domain name pointer
	RecordTypeHINFO                   // 13 - Host information
	RecordTypeMINFO                   // 14 - Mailbox or mail list information
	RecordTypeMX                      // 15 - Mail exchange
	RecordTypeTXT                     // 16 - Text strings

	RecordTypeAAAA  RecordType = 28  // IPv6 address
	RecordTypeSRV   RecordType = 33  // A SRV RR for locating service.
	RecordTypeOPT   RecordType = 41  // An OPT pseudo-RR (sometimes called a meta-RR)
	RecordTypeAXFR  RecordType = 252 // A request for a transfer of an entire zone
	RecordTypeMAILB RecordType = 253 // A request for mailbox-related records (MB, MG or MR)
	RecordTypeMAILA RecordType = 254 // A request for mail agent RRs (Obsolete - see MX)
	RecordTypeALL   RecordType = 255 // A request for all records
)

// RecordTypes contains a mapping between string representation of DNS record
// type with their numeric value, ordered by key alphabetically.
var RecordTypes = map[string]RecordType{
	"A":     RecordTypeA,
	"AAAA":  RecordTypeAAAA,
	"ALL":   RecordTypeALL,
	"AXFR":  RecordTypeAXFR,
	"CNAME": RecordTypeCNAME,
	"HINFO": RecordTypeHINFO,
	"MAILA": RecordTypeMAILA,
	"MAILB": RecordTypeMAILB,
	"MB":    RecordTypeMB,
	"MD":    RecordTypeMD,
	"MF":    RecordTypeMF,
	"MG":    RecordTypeMG,
	"MINFO": RecordTypeMINFO,
	"MR":    RecordTypeMR,
	"MX":    RecordTypeMX,
	"NS":    RecordTypeNS,
	"NULL":  RecordTypeNULL,
	"OPT":   RecordTypeOPT,
	"PTR":   RecordTypePTR,
	"SOA":   RecordTypeSOA,
	"SRV":   RecordTypeSRV,
	"TXT":   RecordTypeTXT,
	"WKS":   RecordTypeWKS,
}

// RecordTypeNames contains mapping between record type and and their string
// representation, ordered alphabetically.
var RecordTypeNames = map[RecordType]string{
	RecordTypeA:     "A",
	RecordTypeAAAA:  "AAAA",
	RecordTypeALL:   "ALL",
	RecordTypeAXFR:  "AXFR",
	RecordTypeCNAME: "CNAME",
	RecordTypeHINFO: "HINFO",
	RecordTypeMAILA: "MAILA",
	RecordTypeMAILB: "MAILB",
	RecordTypeMB:    "MB",
	RecordTypeMD:    "MD",
	RecordTypeMF:    "MF",
	RecordTypeMG:    "MG",
	RecordTypeMINFO: "MINFO",
	RecordTypeMR:    "MR",
	RecordTypeMX:    "MX",
	RecordTypeNS:    "NS",
	RecordTypeNULL:  "NULL",
	RecordTypeOPT:   "OPT",
	RecordTypePTR:   "PTR",
	RecordTypeSOA:   "SOA",
	RecordTypeSRV:   "SRV",
	RecordTypeTXT:   "TXT",
	RecordTypeWKS:   "WKS",
}

// RecordTypeFromAddress return RecordTypeA or RecordTypeAAAA if addr is valid
// IPv4 or IPv6 address, respectively, otherwise it will return 0.
func RecordTypeFromAddress(addr []byte) (rtype RecordType) {
	var (
		ip net.IP = net.ParseIP(string(addr))

		ipv4 net.IP
		ipv6 net.IP
	)

	if ip != nil {
		ipv4 = ip.To4()
		if ipv4 != nil {
			return RecordTypeA
		}
		ipv6 = ip.To16()
		if ipv6 != nil {
			return RecordTypeAAAA
		}
	}
	return 0
}
