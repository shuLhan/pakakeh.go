// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Program gocheck implement go static analysis using [Analyzer] that are not
// included in the default go vet.
// See package [lib/goanalysis] for more information.
//
// [Analyzer]: https://pkg.go.dev/golang.org/x/tools/go/analysis#hdr-Analyzer
// [lib/goanalysis]: https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/goanalysis/
package main

import "git.sr.ht/~shulhan/pakakeh.go/lib/goanalysis"

func main() {
	goanalysis.Check()
}
