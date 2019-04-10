// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewUDPClientPool(t *testing.T) {
	cases := []struct {
		desc   string
		ns     []string
		expErr string
	}{{
		desc:   "With empty name servers",
		expErr: "udp: UDPClientPool: no name servers defined",
	}, {
		desc: "With one invalid name server",
		ns: []string{
			testServerAddress,
			"notipaddress",
		},
		expErr: "dns: invalid address 'notipaddress'",
	}, {
		desc: "With single name server",
		ns: []string{
			testServerAddress,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ucp, err := NewUDPClientPool(c.ns)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		var wg sync.WaitGroup
		qname := []byte("kilabit.info")
		for x := 0; x < 10; x++ {
			wg.Add(1)
			go func() {
				cl := ucp.Get()
				msg, err := cl.Lookup(false, QueryTypeA, QueryClassIN, qname)
				if err != nil {
					t.Log("Lookup error: ", err.Error())
				}
				t.Logf("Lookup response: %+v\n", msg.Header)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
