// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"
)

// # Before 2021-11-14
//
// $ go test -benchmem -bench=MessageHeader_pack -memprofile testdata/pprof/MessageHeader_unpack.mem.old > testdata/bench/MessageHeader_pack.old
//
// goos: linux
// goarch: amd64
// pkg: github.com/shuLhan/share/lib/dns
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
// BenchmarkMessageHeader_pack-8           19629273                66.25 ns/op           32 B/op          3 allocs/op
//
// # After
//
// $ go test -benchmem -bench=MessageHeader_pack -memprofile testdata/pprof/MessageHeader_unpack.mem.new > testdata/bench/MessageHeader_pack.new
//
// goos: linux
// goarch: amd64
// pkg: github.com/shuLhan/share/lib/dns
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
// BenchmarkMessageHeader_pack-8           54183505                21.66 ns/op           16 B/op          1 allocs/op
func BenchmarkMessageHeader_pack(b *testing.B) {
	hdr := &MessageHeader{
		ID:      0xABCD,
		Op:      OpCodeQuery,
		IsAA:    true,
		IsRD:    true,
		QDCount: 1,
		ANCount: 4,
		NSCount: 1,
		ARCount: 1,
	}

	for x := 0; x < b.N; x++ {
		_ = hdr.pack()
	}
}

// $ go test -benchmem -bench=MessageHeader_unpack -memprofile testdata/pprof/MessageHeader_unpack.mem.new > testdata/bench/MessageHeader_unpack.new
//
// goos: linux
// goarch: amd64
// pkg: github.com/shuLhan/share/lib/dns
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
// BenchmarkMessageHeader_unpack-8         310460067                3.848 ns/op           0 B/op          0 allocs/op
func BenchmarkMessageHeader_unpack(b *testing.B) {
	hdr := &MessageHeader{}
	packet := []byte{
		0xab, 0xcd,
		0x85, 0x00,
		0x00, 0x01,
		0x00, 0x04,
		0x00, 0x01,
		0x00, 0x01,
	}

	for x := 0; x < b.N; x++ {
		hdr.unpack(packet)
	}
}
