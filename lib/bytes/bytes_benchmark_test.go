// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"bytes"
	"testing"
)

// go version go1.11.5 linux/amd64
//
// goos: linux
// goarch: amd64
// pkg: github.com/shuLhan/share/lib/bytes
// BenchmarkEqual-4                200000000                7.49 ns/op            0 B/op          0 allocs/op
// BenchmarkCompare-4              200000000                6.88 ns/op            0 B/op          0 allocs/op
func BenchmarkEqual(b *testing.B) {
	s1 := []byte("1234567890123456789012345678901234567890")
	s2 := []byte("1234567890123456789012345678901234567890")
	for x := 0; x < b.N; x++ {
		bytes.Equal(s1, s2)
	}
}

func BenchmarkCompare(b *testing.B) {
	s1 := []byte("1234567890123456789012345678901234567890")
	s2 := []byte("1234567890123456789012345678901234567890")
	for x := 0; x < b.N; x++ {
		bytes.Compare(s1, s2)
	}
}
