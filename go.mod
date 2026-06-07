// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

module kilabit.info/pakakeh.go

go 1.26.0

require (
	golang.org/x/crypto v0.52.0
	golang.org/x/exp v0.0.0-20260603202125-055de637280b
	golang.org/x/net v0.55.0
	golang.org/x/sys v0.45.0
	golang.org/x/term v0.43.0
	golang.org/x/tools v0.45.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

replace golang.org/x/crypto => kilabit.info/go-x-crypto v0.52.1-0.20260607115909-f59b92e8569b

//replace golang.org/x/term => ../../../golang.org/x/term
