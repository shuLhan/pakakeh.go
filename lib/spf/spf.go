// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

// Package spf implement Sender Policy Framework (SPF) per RFC 7208.
package spf

import (
	"log"
	"net"
	"strings"

	libdns "git.sr.ht/~shulhan/pakakeh.go/lib/dns"
)

// List of possible result code as described in RFC 7208 section 2.6.
const (
	// A "pass" result is an explicit statement that the client is
	// authorized to inject mail with the given identity.
	ResultCodePass byte = iota

	// A result of "none" means either,
	// (a) no syntactically valid DNS domain name was extracted from the
	//     SMTP session that could be used as the one to be authorized, or
	// (b) no SPF records were retrieved from the DNS.
	ResultCodeNone

	// A "neutral" result means the ADMD has explicitly stated that it is
	// not asserting whether the IP address is authorized.
	ResultCodeNeutral

	// A "fail" result is an explicit statement that the client is not
	// authorized to use the domain in the given identity.
	ResultCodeFail

	// A "softfail" result is a weak statement by the publishing ADMD that
	// the host is probably not authorized.
	// It has not published a stronger, more definitive policy that
	// results in a "fail".
	ResultCodeSoftfail

	// A "temperror" result means the SPF verifier encountered a transient
	// (generally DNS) error while performing the check.
	// A later retry may // succeed without further DNS operator action.
	ResultCodeTempError

	// A "permerror" result means the domain’s published records could not
	// be correctly interpreted.
	// This signals an error condition that definitely requires DNS
	// operator intervention to be resolved.
	ResultCodePermError
)

var (
	dnsClient       *libdns.UDPClient
	defSystemResolv = ""
	defNameserver   = "1.1.1.1"
)

func init() {
	var err error

	ns := libdns.GetSystemNameServers(defSystemResolv)
	if len(ns) == 0 {
		ns = append(ns, defNameserver)
	}

	dnsClient, err = libdns.NewUDPClient(ns[0])
	if err != nil {
		log.Fatal("spf: init: " + err.Error())
	}
}

// CheckHost fetches SPF records, parses them, and evaluates them to determine
// whether a particular host is or is not permitted to send mail with a given
// identity.
func CheckHost(ip net.IP, domain, sender, hostname string) (result *Result) {
	at := strings.Index(sender, "@")
	if at == -1 {
		sender = "postmaster@" + sender
	}

	result = newResult(ip, domain, sender, hostname)
	if result.Code != 0 {
		return result
	}

	result.lookup()
	if result.Code != 0 {
		return result
	}

	result.evaluateSPFRecord()
	if result.Code != 0 {
		return result
	}

	return result
}
