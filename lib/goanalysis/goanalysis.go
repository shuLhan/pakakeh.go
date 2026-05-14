// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

// Package goanalysis implement go static analysis using
// [Analyzer] that are not included in the default "go vet", but included in
// the [passes] directory, including
//
//   - atomicalign: checks for non-64-bit-aligned arguments to sync/atomic
//     functions.
//   - deepequalerrors: checks for the use of reflect.DeepEqual with error
//     values.
//   - fieldalignment: detects structs that would use less memory if their
//     fields were sorted.
//   - nilness: reports errors such as nil pointer dereferences and degenerate
//     nil pointer comparisons.
//   - reflectvaluecompare: checks for accidentally using == or
//     [reflect.DeepEqual] to compare reflect.Value values.
//   - shadow: checks for shadowed variables.
//   - sortslice: checks for calls to sort.Slice that do not use a slice type
//     as first argument.
//   - unusedwrite: checks for unused writes to the elements of a struct or
//     array object.
//   - usesgenerics: checks for usage of generic features added in Go 1.18.
//
// [Analyzer]: https://pkg.go.dev/golang.org/x/tools/go/analysis#hdr-Analyzer
// [passes]: https://pkg.go.dev/golang.org/x/tools/go/analysis/passes
package goanalysis

import (
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
)

// Check run the static analysis.
// This function is not mean to be call directly, but used in the main func.
func Check() {
	multichecker.Main(
		atomicalign.Analyzer,
		deepequalerrors.Analyzer,
		fieldalignment.Analyzer,
		nilness.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		sortslice.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	)
}
