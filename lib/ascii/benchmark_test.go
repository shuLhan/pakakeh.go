// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ascii

import (
	"bytes"
	"testing"
)

func BenchmarkToLowerStd(b *testing.B) {
	randomInput256 := Random([]byte(HexaLetters), 256)

	in := make([]byte, len(randomInput256))
	copy(in, randomInput256)

	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		bytes.ToLower(in)
	}
}

func BenchmarkToLower(b *testing.B) {
	randomInput256 := Random([]byte(HexaLetters), 256)

	in := make([]byte, len(randomInput256))
	copy(in, randomInput256)

	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		ToLower(&in)
		copy(in, randomInput256)
	}
}

//
// Output of above benchmarks,
//
// goos: linux
// goarch: amd64
// pkg: github.com/shuLhan/share/lib/ascii
// BenchmarkToLowerStd-4            2066588               563 ns/op             256 B/op          1 allocs/op
// BenchmarkToLower-4               5476693               213 ns/op               0 B/op          0 allocs/op
// PASS
// ok      github.com/shuLhan/share/lib/ascii      3.149s
//
