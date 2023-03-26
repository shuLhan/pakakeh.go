// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "fmt"

func ExampleResolvConf_PopulateQuery() {
	var (
		resconf = &ResolvConf{
			Domain: "internal",
			Search: []string{"my.internal"},
			NDots:  1,
		}
		queries []string
	)

	queries = resconf.PopulateQuery("vpn")
	fmt.Println(queries)
	queries = resconf.PopulateQuery("a.machine")
	fmt.Println(queries)
	//Output:
	//[vpn vpn.my.internal]
	//[a.machine]
}
