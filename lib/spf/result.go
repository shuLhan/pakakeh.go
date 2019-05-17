// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

import (
	"bytes"
	"fmt"
	"net"

	libdns "github.com/shuLhan/share/lib/dns"
	libnet "github.com/shuLhan/share/lib/net"
)

//
// Result contains the output of CheckHost function.
//
type Result struct {
	IP       net.IP
	Domain   []byte
	Sender   []byte
	Hostname []byte
	Code     byte
	Err      string

	senderLocal  []byte
	senderDomain []byte
	rdata        []byte // rdata contains raw DNS resource record data that have SPF record.
	dirs         []*directive
	mods         []*modifier
}

//
// newResult initialize new SPF result on single domain.
//
func newResult(ip net.IP, domain, sender, hostname string) (r *Result) {
	bsender := []byte(sender)

	at := bytes.Index(bsender, []byte{'@'})
	r = &Result{
		IP:           ip,
		Domain:       []byte(domain),
		Sender:       []byte(sender),
		Hostname:     []byte(hostname),
		senderLocal:  bsender[:at],
		senderDomain: bsender[at+1:],
	}

	if !libnet.IsHostnameValid(r.Domain, true) {
		r.Code = ResultCodeNone
		r.Err = "invalid domain name"
		return
	}

	return
}

//
// Error return the string representation of the result as error message.
//
func (r *Result) Error() string {
	return fmt.Sprintf("spf: %q %s", r.Domain, r.Err)
}

//
// lookup the TXT record that contains SPF record on domain name.
//
func (r *Result) lookup() {
	var (
		dnsMsg *libdns.Message
		err    error
		txts   []*libdns.ResourceRecord
	)

	dnsMsg, err = dnsClient.Lookup(true, libdns.QueryTypeTXT,
		libdns.QueryClassIN, r.Domain)
	if err != nil {
		r.Code = ResultCodeTempError
		r.Err = err.Error()
		return
	}

	switch dnsMsg.Header.RCode {
	case libdns.RCodeOK:
	case libdns.RCodeErrName:
		r.Code = ResultCodeNone
		r.Err = "domain name does not exist"
		return
	default:
		r.Code = ResultCodeTempError
		r.Err = "server failure"
		return
	}

	txts = dnsMsg.FilterAnswers(libdns.QueryTypeTXT)
	if len(txts) == 0 {
		r.Code = ResultCodeNone
		r.Err = "no SPF record found"
		return
	}

	var found int

	for x := 0; x < len(txts); x++ {
		rdata, ok := txts[x].RData().([]byte)
		if !ok {
			continue
		}

		if bytes.HasPrefix(rdata, []byte("v=spf1")) {
			found++
			if found == 1 {
				r.rdata = rdata
			}
		}
	}
	if found == 0 {
		r.Code = ResultCodeNone
		r.Err = "no SPF record found"
		return
	}
	if found > 1 {
		r.Code = ResultCodePermError
		r.Err = "multiple SPF records found"
		return
	}

	r.rdata = bytes.ToLower(r.rdata)
}

//
// evaluateSPFRecord parse and evaluate each directive with its modifiers
// in the SPF record.
//
func (r *Result) evaluateSPFRecord(rdata []byte) {
	terms := bytes.Fields(rdata)

	// Ignore the first field "v=spf1".

	for x := 1; x < len(terms); x++ {
		dir := r.parseDirective(terms[x])
		if dir != nil {
			continue
		}
		r.dirs = append(r.dirs, dir)
	}
}

//
// parseDirective parse directive from single term.
// It will return non-nil if term is a directive, otherwise it will return
// nil.
//
func (r *Result) parseDirective(term []byte) (dir *directive) {
	var (
		qual int = -1
		err  error
	)

	switch term[0] {
	case '+':
		qual = qualifierPass
	case '-':
		qual = qualifierFail
	case '~':
		qual = qualifierSoftfail
	case '?':
		qual = qualifierNeutral
	}

	if qual == -1 {
		qual = qualifierPass
	} else {
		term = term[1:]
	}

	// Try to split the term to get the mechanism and domain-spec.
	kv := bytes.Split(term, []byte{':'})

	switch {
	case bytes.Equal(kv[0], []byte("all")):
		dir = &directive{
			qual: qual,
			mech: mechanismAll,
		}
		return dir

	case bytes.Equal(kv[0], []byte("include")):
		dir = &directive{
			qual: qual,
			mech: mechanismInclude,
		}
		if len(kv) >= 2 {
			dir.value, err = macroExpand(r, mechanismInclude, kv[1])
			if err != nil {
				r.Code = ResultCodePermError
				r.Err = err.Error()
				return nil
			}
		}
		return dir
	}

	return nil
}
