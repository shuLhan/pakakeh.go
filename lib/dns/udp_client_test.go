// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestUDPClientLookup(t *testing.T) {
	type testCase struct {
		exp            *Message
		desc           string
		qst            MessageQuestion
		allowRecursion bool
	}

	var (
		cases []testCase
		c     testCase
		cl    *UDPClient
		got   *Message
		err   error
	)

	cl, err = NewUDPClient(testServerAddress)
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
				ID:      8,
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
				ID:      9,
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
				ID:      10,
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
				ID:      11,
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
	}, {
		desc: "IsRD:true QType:AAAA RClass:IN QName:kilabit.info",
		qst: MessageQuestion{
			Name: "kilabit.info",
			Type: RecordTypeAAAA,
		},
		allowRecursion: true,
	}, {
		desc: "IsRD:true QType:A RClass:IN QName:random",
		qst: MessageQuestion{
			Name: "random",
			Type: RecordTypeA,
		},
		allowRecursion: true,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		got, err = cl.Lookup(c.qst, c.allowRecursion)
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
