// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

module git.sr.ht/~shulhan/pakakeh.go

go 1.25.0

require (
	golang.org/x/crypto v0.48.0
	golang.org/x/exp v0.0.0-20260209203927-2842357ff358
	golang.org/x/net v0.50.0
	golang.org/x/sys v0.41.0
	golang.org/x/term v0.40.0
	golang.org/x/tools v0.42.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
)

//replace golang.org/x/crypto => ../go-x-crypto

//replace golang.org/x/term => ../../../golang.org/x/term
