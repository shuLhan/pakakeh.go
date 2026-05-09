// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

module git.sr.ht/~shulhan/pakakeh.go

go 1.26.0

require (
	golang.org/x/crypto v0.50.0
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90
	golang.org/x/net v0.53.0
	golang.org/x/sys v0.44.0
	golang.org/x/term v0.43.0
	golang.org/x/tools v0.43.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

replace golang.org/x/crypto => git.sr.ht/~shulhan/go-x-crypto v0.51.1-0.20260509111418-c27adc710d5b

//replace golang.org/x/term => ../../../golang.org/x/term
