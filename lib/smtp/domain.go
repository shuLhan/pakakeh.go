// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Domain contains a host name and list of accounts.
//
type Domain struct {
	Name     string
	Accounts []*Account
}

//
// NewDomain create new domain with single main user, "postmaster".
//
func NewDomain(name string) (domain *Domain) {
	accPostmaster := NewAccount("Postmaster", "postmaster", "")

	domain = &Domain{
		Name: name,
		Accounts: []*Account{
			accPostmaster,
		},
	}

	return
}
