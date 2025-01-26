// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package goanalysis implement go static analysis using
// [Analyzer] that are not included in the default "go vet", but included in
// the [passes] directory, including
//
//   - fieldalignment: detects structs that would use less memory if their
//     fields were sorted.
//   - nilness: inspects the control-flow graph of an SSA function and reports
//     errors such as nil pointer dereferences and degenerate nil pointer
//     comparisons.
//   - reflectvaluecompare: checks for accidentally using == or
//     [reflect.DeepEqual] to compare reflect.Value values.
//   - shadow: checks for shadowed variables.
//   - sortslice: checks for calls to sort.Slice that do not use a slice type
//     as first argument.
//   - unusedwrite: checks for unused writes to the elements of a struct or
//     array object.
//   - waitgroup: detects simple misuses of sync.WaitGroup.
//
// [Analyzer]: https://pkg.go.dev/golang.org/x/tools/go/analysis#hdr-Analyzer
// [passes]: https://pkg.go.dev/golang.org/x/tools/go/analysis/passes
package goanalysis

import (
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
)

// Check run the static analysis.
// This function is not mean to be call directly, but used in the main func.
func Check() {
	multichecker.Main(
		fieldalignment.Analyzer,
		nilness.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		sortslice.Analyzer,
		unusedwrite.Analyzer,
		waitgroup.Analyzer,
	)
}
