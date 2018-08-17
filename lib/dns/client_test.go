// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	_ "github.com/shuLhan/share/lib/test"
)

func TestClientLookup(t *testing.T) {
	cl, err := NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		qtype  QueryType
		qclass uint16
		qname  []byte
	}{{
		desc:   "QType:A QClass:IN QName:kilabit.info",
		qtype:  QueryTypeA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
	}, {
		desc:   "QType:SOA QClass:IN QName:kilabit.info",
		qtype:  QueryTypeSOA,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
	}, {
		desc:   "QType:TXT QClass:IN QName:kilabit.info",
		qtype:  QueryTypeTXT,
		qclass: QueryClassIN,
		qname:  []byte("kilabit.info"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, err := cl.Lookup(c.qtype, c.qclass, c.qname)
		if err != nil {
			t.Fatal(err)
		}
	}
}
