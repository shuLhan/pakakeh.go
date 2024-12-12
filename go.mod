// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

module git.sr.ht/~shulhan/pakakeh.go

go 1.22.0

require (
	golang.org/x/crypto v0.31.0
	golang.org/x/net v0.32.0
	golang.org/x/sys v0.28.0
	golang.org/x/term v0.27.0
	golang.org/x/tools v0.28.0
)

require (
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
)

//replace golang.org/x/crypto => ../go-x-crypto

//replace golang.org/x/term => ../../../golang.org/x/term
