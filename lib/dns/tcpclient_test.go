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
		desc   string
		qtype  uint16
		qclass uint16
		qname  []byte
		exp    *Message
	}{{
		desc:   "QType:A QClass:IN QName:kilabit.info",
		qtype:  QueryTypeA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		exp: &Message{
			Header: &SectionHeader{
				ID:      5,
				QDCount: 1,
				ANCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				rdlen: 4,
				Text: &RDataText{
					Value: []byte("127.0.0.1"),
				},
			}},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}, {
		desc:   "QType:SOA QClass:IN QName:kilabit.info",
		qtype:  QueryTypeSOA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		exp: &Message{
			Header: &SectionHeader{
				ID:      6,
				QDCount: 1,
				ANCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				SOA: &RDataSOA{
					MName:   []byte("kilabit.info"),
					RName:   []byte("admin.kilabit.info"),
					Serial:  20180832,
					Refresh: 3600,
					Retry:   60,
					Expire:  3600,
					Minimum: 3600,
				},
			}},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}, {
		desc:   "QType:TXT QClass:IN QName:kilabit.info",
		qtype:  QueryTypeTXT,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		exp: &Message{
			Header: &SectionHeader{
				ID:      7,
				QDCount: 1,
				ANCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					Value: []byte("This is a test server"),
				},
			}},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := cl.Lookup(c.qtype, c.qclass, c.qname)
		if err != nil {
			t.Fatal(err)
		}

		c.exp.Header.ID = getID()

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.Packet, got.Packet, true)
	}
}
