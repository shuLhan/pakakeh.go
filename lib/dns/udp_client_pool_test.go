// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewUDPClientPool(t *testing.T) {
	type testCase struct {
		desc   string
		expErr string
		ns     []string
	}

	var (
		cases []testCase
		c     testCase

		clPool *UDPClientPool
		wg     sync.WaitGroup

		qname string
		err   error
		x     int
	)

	cases = []testCase{{
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

	for _, c = range cases {
		t.Log(c.desc)

		clPool, err = NewUDPClientPool(c.ns)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		qname = "kilabit.info"
		for x = 0; x < 10; x++ {
			wg.Add(1)
			go func() {
				var (
					cl = clPool.Get()
					q  = MessageQuestion{
						Name: qname,
					}

					err error
				)
				_, err = cl.Lookup(q, false)
				if err != nil {
					t.Log("Lookup error: ", err.Error())
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
