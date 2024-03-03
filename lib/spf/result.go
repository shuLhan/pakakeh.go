// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dns"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

// Result contains the output of CheckHost function.
type Result struct {
	IP net.IP // The IP address of sender.

	Err string

	Domain   []byte // The domain address of sender from SMTP EHLO or MAIL FROM command.
	Sender   []byte // The email address of sender.
	Hostname []byte

	senderLocal  []byte // The local part of sender.
	senderDomain []byte // The domain part of sender.
	terms        []byte // terms contains raw DNS RR that have SPF record.
	dirs         []*directive
	mods         []*modifier

	Code byte // Result of check host.
}

// newResult initialize new SPF result on single domain.
func newResult(ip net.IP, domain, sender, hostname string) (result *Result) {
	bsender := []byte(sender)

	at := bytes.Index(bsender, []byte{'@'})
	result = &Result{
		IP:           ip,
		Domain:       []byte(domain),
		Sender:       []byte(sender),
		Hostname:     []byte(hostname),
		senderLocal:  bsender[:at],
		senderDomain: bsender[at+1:],
	}

	if !libnet.IsHostnameValid(result.Domain, true) {
		result.Code = ResultCodeNone
		result.Err = "invalid domain name"
		return
	}

	return
}

// Error return the string representation of the result as error message.
func (result *Result) Error() string {
	return fmt.Sprintf("spf: %q %s", result.Domain, result.Err)
}

// lookup the TXT record that contains SPF record on domain name.
func (result *Result) lookup() {
	var (
		dnsMsg *dns.Message
		err    error
		txts   []dns.ResourceRecord
		q      = dns.MessageQuestion{
			Name: string(result.Domain),
			Type: dns.RecordTypeTXT,
		}
	)
	dnsMsg, err = dnsClient.Lookup(q, true)
	if err != nil {
		result.Code = ResultCodeTempError
		result.Err = err.Error()
		return
	}

	switch dnsMsg.Header.RCode {
	case dns.RCodeOK:
		// NOOP.
	case dns.RCodeErrName:
		result.Code = ResultCodeNone
		result.Err = "domain name does not exist"
		return
	case dns.RCodeErrFormat, dns.RCodeErrServer, dns.RCodeNotImplemented, dns.RCodeRefused:
		result.Code = ResultCodeTempError
		result.Err = "server failure"
		return
	}

	txts = dnsMsg.FilterAnswers(dns.RecordTypeTXT)
	if len(txts) == 0 {
		result.Code = ResultCodeNone
		result.Err = "no SPF record found"
		return
	}

	var found int

	for x := 0; x < len(txts); x++ {
		rdata, ok := txts[x].Value.(string)
		if !ok {
			continue
		}

		if strings.HasPrefix(rdata, "v=spf1") {
			found++
			if found == 1 {
				result.terms = []byte(rdata)
			}
		}
	}
	if found == 0 {
		result.Code = ResultCodeNone
		result.Err = "no SPF record found"
		return
	}
	if found > 1 {
		result.Code = ResultCodePermError
		result.Err = "multiple SPF records found"
		return
	}

	result.terms = bytes.ToLower(result.terms)
}

// evaluateSPFRecord parse and evaluate each directive with its modifiers
// in the SPF record.
//
//	terms            = *( 1*SP ( directive / modifier ) )
//
//	directive        = [ qualifier ] mechanism
//	qualifier        = "+" / "-" / "?" / "~"
//	mechanism        = ( all / include
//	                 / a / mx / ptr / ip4 / ip6 / exists )
//	modifier         = redirect / explanation / unknown-modifier
//	unknown-modifier = name "=" macro-string
//	                 ; where name is not any known modifier
//
//	name             = ALPHA *( ALPHA / DIGIT / "-" / "_" / "." )
func (result *Result) evaluateSPFRecord() {
	terms := bytes.Fields(result.terms)

	// Ignore the first field "v=spf1".

	for x := 1; x < len(terms); x++ {
		dir := result.parseDirective(terms[x])
		if dir != nil {
			result.dirs = append(result.dirs, dir)

			// Mechanisms after "all" will never be tested and MUST be
			// ignored -- RFC 7208 section 5.1.
			if dir.mech == mechanismAll {
				return
			}
			continue
		}

		if result.Code != 0 {
			return
		}

		mod := result.parseModifier(terms[x])
		if mod == nil {
			return
		}

		result.mods = append(result.mods, mod)
	}
}

// parseDirective parse directive from single term.
// It will return non-nil if term is a directive, otherwise it will return
// nil.
func (result *Result) parseDirective(term []byte) (dir *directive) {
	var (
		qual byte
		err  error
	)

	switch term[0] {
	case qualifierPass, qualifierNeutral, qualifierSoftfail, qualifierFail:
		qual = term[0]
		term = term[1:]
	default:
		qual = qualifierPass
	}

	// Try to split the term to get the mechanism and domain-spec.
	kv := bytes.Split(term, []byte{':'})

	mech := string(kv[0])

	switch mech {
	case mechanismAll:
		dir = &directive{
			qual: qual,
			mech: mech,
		}
		return dir

	case mechanismInclude:
		dir = &directive{
			qual: qual,
			mech: mech,
		}
		if len(kv) >= 2 {
			dir.value, err = macroExpand(result, mechanismInclude, kv[1])
			if err != nil {
				result.Code = ResultCodePermError
				result.Err = err.Error()
				return nil
			}
		}
		return dir

	case mechanismA:

	case mechanismMx:

	case mechanismPtr:

	case mechanismIP4:

	case mechanismIP6:

	case mechanismExist:
	}

	return nil
}

func (result *Result) parseModifier(term []byte) (mod *modifier) {
	kv := bytes.Split(term, []byte{'='})

	mod = &modifier{
		name: string(kv[0]),
	}
	if len(kv) >= 2 {
		mod.value = string(kv[1])
	}

	switch mod.name {
	case modifierExp:
		return mod
	case modifierRedirect:
		return mod
	}
	return mod
}
