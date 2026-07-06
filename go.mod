// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

module kilabit.info/pakakeh.go

go 1.26.0

require (
	golang.org/x/crypto v0.53.0
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976
	golang.org/x/net v0.56.0
	golang.org/x/sys v0.46.0
	golang.org/x/term v0.44.0
	golang.org/x/tools v0.47.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
)

replace golang.org/x/crypto => kilabit.info/go-x-crypto v0.53.1-0.20260706045628-1aeb5abefb85

//replace golang.org/x/term => ../../../golang.org/x/term
