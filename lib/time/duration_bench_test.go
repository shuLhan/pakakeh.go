// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

// Before,
//
//	BenchmarkParseDuration-4  5000000  332 ns/op  32 B/op  4 allocs/op
func BenchmarkParseDuration(b *testing.B) {
	exp := time.Duration(1.5 * float64(Week))

	for x := 0; x < b.N; x++ {
		got, err := ParseDuration("1w0.5w")
		if err != nil {
			b.Fatal(err)
		}

		test.Assert(b, "duration", exp, got)
	}
}
