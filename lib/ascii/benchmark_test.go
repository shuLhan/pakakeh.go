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
		ToLower(in)
		copy(in, randomInput256) // Copy original input back.
	}
}

/****
Output of above benchmarks,

=== go version devel go1.18-d38f31d805

Wed 13 Oct 17:45:42 UTC 2021

goos: linux
goarch: amd64
pkg: github.com/shuLhan/share/lib/ascii
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkToLowerStd-8            3683070               328.9 ns/op           256 B/op          1 allocs/op
BenchmarkToLower-8               5221684               232.0 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/shuLhan/share/lib/ascii      2.990s

=== go version ???

goos: linux
goarch: amd64
pkg: github.com/shuLhan/share/lib/ascii
BenchmarkToLowerStd-4            2066588               563 ns/op             256 B/op          1 allocs/op
BenchmarkToLower-4               5476693               213 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/shuLhan/share/lib/ascii      3.149s

****/
