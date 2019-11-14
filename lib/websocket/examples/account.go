// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package examples

//
// Account represent an example of internal user in the system.
//
type Account struct {
	ID   int64
	Name string
	Key  string // The Key to authenticate user during handshake.
}

//
// List of user's account in the system.
//
var Users map[int64]*Account = map[int64]*Account{
	1: {
		ID:   1,
		Name: "Groot",
		Key:  "iamgroot",
	},
	2: {
		ID:   2,
		Name: "Thanos",
		Key:  "thanosdidnothingwrong",
	},
	3: {
		ID:   3,
		Name: "Hulk",
		Key:  "arrrr",
	},
}
