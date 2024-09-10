// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package examples

// Account represent an example of internal user in the system.
type Account struct {
	Name string
	Key  string // The Key to authenticate user during handshake.
	ID   int64
}

// Users contain list of user's account in the system.
var Users = map[int64]*Account{
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
	4: {
		ID:   4,
		Name: `Ironman`,
		Key:  `pewpew`,
	},
}
