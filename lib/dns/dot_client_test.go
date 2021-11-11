// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestDoTClient_Lookup(t *testing.T) {
	cl, err := NewDoTClient(testDoTServerAddress, true)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc           string
		allowRecursion bool
		qtype          uint16
		qclass         uint16
		qname          string
		exp            *Message
	}{{
		desc:   "QType:A QClass:IN QName:kilabit.info",
		qtype:  QueryTypeA,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      5,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  QueryTypeA,
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
		qtype:  QueryTypeSOA,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      6,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  QueryTypeSOA,
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
		qtype:  QueryTypeTXT,
		qclass: QueryClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: SectionHeader{
				ID:      7,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  "kilabit.info",
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
				TTL:   3600,
				Value: "This is a test server",
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Lookup(c.allowRecursion, c.qtype, c.qclass, c.qname)
		if err != nil {
			t.Fatal(err)
		}

		c.exp.Header.ID = got.Header.ID

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.packet, got.packet)
	}
}
