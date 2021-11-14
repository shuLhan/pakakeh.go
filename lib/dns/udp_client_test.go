// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestUDPClientLookup(t *testing.T) {
	cl, err := NewUDPClient(testServerAddress)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc           string
		allowRecursion bool
		rtype          RecordType
		qclass         uint16
		qname          string
		exp            *Message
	}{{
		desc:   "QType:A QClass:IN QName:kilabit.info",
		rtype:  RecordTypeA,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      8,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				rdlen: 4,
				Value: "127.0.0.1",
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}, {
		desc:   "QType:SOA QClass:IN QName:kilabit.info",
		rtype:  RecordTypeSOA,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      9,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataSOA{
					MName:   "kilabit.info",
					RName:   "admin.kilabit.info",
					Serial:  20180832,
					Refresh: 3600,
					Retry:   60,
					Expire:  3600,
					Minimum: 3600,
				},
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}, {
		desc:   "QType:TXT QClass:IN QName:kilabit.info",
		rtype:  RecordTypeTXT,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      10,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeTXT,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeTXT,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "This is a test server",
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}, {
		desc:   "QType:AAAA QClass:IN QName:kilabit.info",
		rtype:  RecordTypeAAAA,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      11,
				IsAA:    false,
				RCode:   RCodeErrServer,
				QDCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeAAAA,
				Class: QueryClassIN,
			},
			Answer:     []ResourceRecord{},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}, {
		desc:           "IsRD:true QType:AAAA QClass:IN QName:kilabit.info",
		allowRecursion: true,
		rtype:          RecordTypeAAAA,
		qclass:         QueryClassIN,
		qname:          "kilabit.info",
	}, {
		desc:           "IsRD:true QType:A QClass:IN QName:random",
		allowRecursion: true,
		rtype:          RecordTypeA,
		qclass:         QueryClassIN,
		qname:          "random",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Lookup(c.allowRecursion, c.rtype, c.qclass, c.qname)
		if err != nil {
			t.Fatal(err)
		}

		if !c.allowRecursion {
			c.exp.Header.ID = getID()

			_, err = c.exp.Pack()
			if err != nil {
				t.Fatal(err)
			}

			test.Assert(t, "packet", c.exp.packet, got.packet)
		} else {
			t.Logf("Got recursive answer: %+v\n", got)
		}
	}
}
