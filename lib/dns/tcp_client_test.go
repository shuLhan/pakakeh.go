// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestTCPClientLookup(t *testing.T) {
	cl, err := NewTCPClient(testServerAddress)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc           string
		allowRecursion bool
		rtype          RecordType
		rclass         RecordClass
		qname          string
		exp            *Message
	}{{
		desc:   "RType:A RClass:IN QName:kilabit.info",
		rtype:  RecordTypeA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      5,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: RecordClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: RecordClassIN,
				TTL:   3600,
				rdlen: 4,
				Value: "127.0.0.1",
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}, {
		desc:   "RType:SOA RClass:IN QName:kilabit.info",
		rtype:  RecordTypeSOA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      6,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeSOA,
				Class: RecordClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeSOA,
				Class: RecordClassIN,
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
		desc:   "RType:TXT RClass:IN QName:kilabit.info",
		rtype:  RecordTypeTXT,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      7,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeTXT,
				Class: RecordClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeTXT,
				Class: RecordClassIN,
				TTL:   3600,
				Value: "This is a test server",
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Lookup(c.allowRecursion, c.rtype, c.rclass, c.qname)
		if err != nil {
			t.Fatal(err)
		}

		c.exp.Header.ID = getID()

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "packet", c.exp.packet, got.packet)
	}
}
