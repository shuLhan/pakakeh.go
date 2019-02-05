// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bytes"
	"testing"
)

// Before:
//
//	BenchmarkReaderReadUntil-4  5000000  247 ns/op  32.35 MB/s  32 B/op  1 allocs/op
//
// After:
//
//	BenchmarkReaderReadUntil-4  5000000  239 ns/op  33.37 MB/s  32 B/op  1 allocs/op
//
func BenchmarkReaderReadUntil(b *testing.B) {
	r := &Reader{}

	seps := []byte{',', '.', '|'}
	terms := []byte{'\n'}

	b.SetBytes(8)
	var (
		tok []byte
		c   byte
	)

	for x := 0; x < b.N; x++ {
		r.Init([]byte("xxx|yyy|zzz|000|111|222"))

		tok, _, _ = r.ReadUntil(seps, terms)
		if !bytes.Equal(tok, []byte("xxx")) {
			b.Fatal("first token not match!")
		}

		tok, _, _ = r.ReadUntil(seps, terms)
		if !bytes.Equal(tok, []byte("yyy")) {
			b.Fatal("second token not match!")
		}

		tok, _, _ = r.ReadUntil(seps, terms)
		if !bytes.Equal(tok, []byte("zzz")) {
			b.Fatal("third token not match!")
		}

		tok, _, _ = r.ReadUntil(seps, terms)
		if !bytes.Equal(tok, []byte("000")) {
			b.Fatal("fourth token not match!")
		}

		tok, _, _ = r.ReadUntil(seps, terms)
		if !bytes.Equal(tok, []byte("111")) {
			b.Fatal("fifth token not match!")
		}

		tok, _, c = r.ReadUntil(seps, terms)
		if !bytes.Equal(tok, []byte("222")) {
			b.Fatal("sixth token not match!")
		}
		if c != 0 {
			b.Fatal("expecting EOB!")
		}
	}
}
