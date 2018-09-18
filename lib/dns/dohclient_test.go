// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestDoHClient_Lookup(t *testing.T) {
	nameserver := "https://127.0.0.1:8443/dns-query"

	cl, err := NewDoHClient(nameserver, true)
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
				ID:      0,
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
				ID:      0,
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
				ID:      0,
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
	}, {
		desc:   "QType:AAAA QClass:IN QName:kilabit.info",
		qtype:  QueryTypeAAAA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		exp: &Message{
			Header: &SectionHeader{
				ID:      0,
				QDCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeAAAA,
				Class: QueryClassIN,
			},
			Answer:     []*ResourceRecord{},
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

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.Packet, got.Packet, true)
	}
}

func TestDoHClient_Post(t *testing.T) {
	nameserver := "https://127.0.0.1:8443/dns-query"

	cl, err := NewDoHClient(nameserver, true)
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
				ID:      0,
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
				ID:      0,
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
				ID:      0,
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
	}, {
		desc:   "QType:AAAA QClass:IN QName:kilabit.info",
		qtype:  QueryTypeAAAA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		exp: &Message{
			Header: &SectionHeader{
				ID:      0,
				QDCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeAAAA,
				Class: QueryClassIN,
			},
			Answer:     []*ResourceRecord{},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		msg := NewMessage()

		msg.Question.Type = c.qtype
		msg.Question.Class = c.qclass
		msg.Question.Name = append(msg.Question.Name, c.qname...)

		_, err := msg.Pack()
		if err != nil {
			t.Fatal("msg.Pack:", err)
		}

		got, err := cl.Post(msg)
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.Packet, got.Packet, true)
	}
}

func TestDoHClient_Get(t *testing.T) {
	nameserver := "https://127.0.0.1:8443/dns-invalid"

	cl, err := NewDoHClient(nameserver, true)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		qtype  uint16
		qclass uint16
		qname  []byte
		exp    *Message
		expErr string
	}{{
		desc:   "QType:A QClass:IN QName:kilabit.info",
		qtype:  QueryTypeA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		expErr: "404 page not found\n",
	}, {
		desc:   "QType:A QClass:IN QName:kilabit.info",
		qtype:  QueryTypeA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
		expErr: "404 page not found\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		msg := NewMessage()

		msg.Question.Type = c.qtype
		msg.Question.Class = c.qclass
		msg.Question.Name = append(msg.Question.Name, c.qname...)

		_, err := msg.Pack()
		if err != nil {
			t.Fatal("msg.Pack:", err)
		}

		got, err := cl.Get(msg)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.Packet, got.Packet, true)
	}
}
