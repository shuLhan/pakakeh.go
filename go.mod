// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

module git.sr.ht/~shulhan/pakakeh.go

go 1.23.4

require (
	golang.org/x/crypto v0.39.0
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b
	golang.org/x/net v0.41.0
	golang.org/x/sys v0.33.0
	golang.org/x/term v0.32.0
	golang.org/x/tools v0.34.0
)

require (
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
)

//replace golang.org/x/crypto => ../go-x-crypto

//replace golang.org/x/term => ../../../golang.org/x/term
