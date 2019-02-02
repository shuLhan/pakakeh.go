// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"testing"
)

func BenchmarkReaderReadUntil(b *testing.B) {
	r := &Reader{}
	r.InitBytes([]byte("xxx yyy zzz | xxx yyy zzz"))

	seps := []byte{',', '.', '|'}
	terms := []byte{'\n'}

	b.SetBytes(8)

	for x := 0; x < b.N; x++ {
		b, isTerm, c := r.ReadUntil(seps, terms)
		_ = b
		_ = isTerm
		_ = c
	}
}
