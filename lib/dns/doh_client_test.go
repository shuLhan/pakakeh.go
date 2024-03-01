// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestDoHClient_Lookup(t *testing.T) {
	type testCase struct {
		exp            *Message
		desc           string
		qst            MessageQuestion
		allowRecursion bool
	}

	var (
		nameserver = "https://127.0.0.1:8443/dns-query"

		cases []testCase
		c     testCase
		got   *Message
		cl    *DoHClient
		err   error
	)

	cl, err = NewDoHClient(nameserver, true)
	if err != nil {
		t.Fatal(err)
	}

	cases = []testCase{{
		desc: "QType:A RClass:IN QName:kilabit.info",
		qst: MessageQuestion{
			Name: "kilabit.info",
		},
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
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
		desc: "QType:SOA RClass:IN QName:kilabit.info",
		qst: MessageQuestion{
			Name: "kilabit.info",
			Type: RecordTypeSOA,
		},
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
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
		desc: "QType:TXT RClass:IN QName:kilabit.info",
		qst: MessageQuestion{
			Name: "kilabit.info",
			Type: RecordTypeTXT,
		},
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
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
	}, {
		desc: "QType:AAAA RClass:IN QName:kilabit.info",
		qst: MessageQuestion{
			Name: "kilabit.info",
			Type: RecordTypeAAAA,
		},
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    false,
				RCode:   RCodeErrServer,
				QDCount: 1,
			},
			Question: MessageQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeAAAA,
				Class: RecordClassIN,
			},
			Answer:     []ResourceRecord{},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		got, err = cl.Lookup(c.qst, c.allowRecursion)
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.packet, got.packet)
	}
}

func TestDoHClient_Post(t *testing.T) {
	type testCase struct {
		exp            *Message
		desc           string
		qname          string
		rtype          RecordType
		rclass         RecordClass
		allowRecursion bool
	}

	var (
		nameserver = "https://127.0.0.1:8443/dns-query"

		cases []testCase
		c     testCase
		cl    *DoHClient
		msg   *Message
		got   *Message
		err   error
	)

	cl, err = NewDoHClient(nameserver, true)
	if err != nil {
		t.Fatal(err)
	}

	cases = []testCase{{
		desc:   "QType:A RClass:IN QName:kilabit.info",
		rtype:  RecordTypeA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
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
		desc:   "QType:SOA RClass:IN QName:kilabit.info",
		rtype:  RecordTypeSOA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
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
		desc:   "QType:TXT RClass:IN QName:kilabit.info",
		rtype:  RecordTypeTXT,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
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
	}, {
		desc:   "QType:AAAA RClass:IN QName:kilabit.info",
		rtype:  RecordTypeAAAA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		exp: &Message{
			Header: MessageHeader{
				ID:      0,
				IsAA:    false,
				RCode:   RCodeErrServer,
				QDCount: 1,
			},
			Question: MessageQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeAAAA,
				Class: RecordClassIN,
			},
			Answer:     []ResourceRecord{},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		},
	}}

	for _, c = range cases {
		t.Log(c.desc)

		msg = NewMessage()

		msg.Header.IsRD = c.allowRecursion
		msg.Question.Type = c.rtype
		msg.Question.Class = c.rclass
		msg.Question.Name = c.qname

		_, err = msg.Pack()
		if err != nil {
			t.Fatal("msg.Pack:", err)
		}

		got, err = cl.Post(msg)
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.packet, got.packet)
	}
}

func TestDoHClient_Get(t *testing.T) {
	type testCase struct {
		exp    *Message
		desc   string
		qname  string
		expErr string
		rtype  RecordType
		rclass RecordClass
	}

	var (
		nameserver = "https://127.0.0.1:8443/dns-invalid"

		cases []testCase
		c     testCase
		cl    *DoHClient
		msg   *Message
		got   *Message
		err   error
	)

	cl, err = NewDoHClient(nameserver, true)
	if err != nil {
		t.Fatal(err)
	}

	cases = []testCase{{
		desc:   "QType:A RClass:IN QName:kilabit.info",
		rtype:  RecordTypeA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		expErr: "Get: 404 page not found\n",
	}, {
		desc:   "QType:A RClass:IN QName:kilabit.info",
		rtype:  RecordTypeA,
		rclass: RecordClassIN,
		qname:  "kilabit.info",
		expErr: "Get: 404 page not found\n",
	}}

	for _, c = range cases {
		t.Log(c.desc)

		msg = NewMessage()

		msg.Question.Type = c.rtype
		msg.Question.Class = c.rclass
		msg.Question.Name = c.qname

		_, err = msg.Pack()
		if err != nil {
			t.Fatal("msg.Pack:", err)
		}

		got, err = cl.Get(msg)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		_, err = c.exp.Pack()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Packet", c.exp.packet, got.packet)
	}
}
