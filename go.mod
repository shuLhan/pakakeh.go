// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

module git.sr.ht/~shulhan/pakakeh.go

go 1.26.0

require (
	golang.org/x/crypto v0.49.0
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90
	golang.org/x/net v0.52.0
	golang.org/x/sys v0.42.0
	golang.org/x/term v0.41.0
	golang.org/x/tools v0.43.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

replace golang.org/x/crypto => git.sr.ht/~shulhan/go-x-crypto v0.49.1-0.20260326231004-e6291d138282

//replace golang.org/x/term => ../../../golang.org/x/term
