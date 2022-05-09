// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

// Domain contains a host name and list of accounts in domain, with optional
// DKIM feature.
type Domain struct {
	dkimOpts *DKIMOptions

	Accounts map[string]*Account
	Name     string
}

// NewDomain create new domain with single main user, "postmaster".
func NewDomain(name string, dkimOpts *DKIMOptions) (domain *Domain) {
	accPostmaster, _ := NewAccount("Postmaster", "postmaster", "", "")

	domain = &Domain{
		Name: name,
		Accounts: map[string]*Account{
			"postmaster": accPostmaster,
		},
		dkimOpts: dkimOpts,
	}

	return
}
