// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net_test

import (
	"bytes"
	"fmt"
	"log"

	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

func ExampleResolvConf_PopulateQuery() {
	var (
		resconf = &libnet.ResolvConf{
			Domain: `internal`,
			Search: []string{`my.internal`},
			NDots:  1,
		}
		queries []string
	)

	queries = resconf.PopulateQuery(`vpn`)
	fmt.Println(queries)
	queries = resconf.PopulateQuery(`a.machine`)
	fmt.Println(queries)

	// Output:
	// [vpn vpn.my.internal]
	// [a.machine]
}

func ExampleResolvConf_WriteTo() {
	var (
		rc = &libnet.ResolvConf{
			Domain:      `internal`,
			Search:      []string{`a.internal`, `b.internal`},
			NameServers: []string{`127.0.0.1`, `192.168.1.1`},
			NDots:       1,
			Timeout:     5,
			Attempts:    3,
			OptMisc: map[string]bool{
				`rotate`: true,
				`debug`:  true,
			},
		}

		bb  bytes.Buffer
		err error
	)

	_, err = rc.WriteTo(&bb)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(bb.String())
	// Output:
	// domain internal
	// search a.internal b.internal
	// nameserver 127.0.0.1
	// nameserver 192.168.1.1
	// options ndots:1
	// options timeout:5
	// options attempts:3
	// options debug
	// options rotate
}
