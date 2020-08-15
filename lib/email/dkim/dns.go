// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"errors"
	"fmt"

	"github.com/shuLhan/share/lib/dns"
)

var dnsClientPool *dns.UDPClientPool

func newDNSClientPool() (err error) {
	var ns []string

	if len(DefaultNameServers) > 0 {
		ns = DefaultNameServers
	} else {
		ns = dns.GetSystemNameServers("")
		if len(ns) == 0 {
			ns = append(ns, "1.1.1.1")
		}
	}

	dnsClientPool, err = dns.NewUDPClientPool(ns)
	if err != nil {
		err = errors.New("dkim: newDNSClientPool: " + err.Error())
		return err
	}

	return nil
}

func lookupDNS(opt QueryOption, dname string) (key *Key, err error) {
	if opt == QueryOptionTXT {
		key, err = lookupDNSTXT(dname)
	}
	return key, err
}

func lookupDNSTXT(dname string) (key *Key, err error) {
	if len(dname) == 0 {
		return nil, nil
	}

	if dnsClientPool == nil {
		err = newDNSClientPool()
		if err != nil {
			return nil, err
		}
	}

	dnsClient := dnsClientPool.Get()

	dnsMsg, err := dnsClient.Lookup(true, dns.QueryTypeTXT,
		dns.QueryClassIN, dname)
	if err != nil {
		dnsClientPool.Put(dnsClient)
		return nil, fmt.Errorf("dkim: LookupKey: %w", err)
	}
	if dnsMsg.Header.RCode != dns.RCodeOK {
		dnsClientPool.Put(dnsClient)
		return nil, fmt.Errorf("dkim: LookupKey: DNS response status: %d",
			dnsMsg.Header.RCode)
	}
	if len(dnsMsg.Answer) == 0 {
		dnsClientPool.Put(dnsClient)
		return nil, fmt.Errorf("dkim: LookupKey: empty answer on '%s'", dname)
	}

	dnsClientPool.Put(dnsClient)

	answers := dnsMsg.FilterAnswers(dns.QueryTypeTXT)
	if len(answers) == 0 {
		return nil, fmt.Errorf("dkim: LookupKey: no TXT record on '%s'", dname)
	}
	if len(answers) != 1 {
		return nil, fmt.Errorf("dkim: LookupKey: multiple TXT records on '%s'", dname)
	}

	txt := answers[0].Value.(string)

	return ParseTXT([]byte(txt), answers[0].TTL)
}
