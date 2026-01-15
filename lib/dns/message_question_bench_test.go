// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package dns

import "testing"

// # 2021-11-15
//
// goos: linux
// goarch: amd64
// pkg: git.sr.ht/~shulhan/pakakeh.go/lib/dns
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
// BenchmarkMessageQuestion_String-8        7138899               168.3 ns/op            56 B/op          3 allocs/op
func BenchmarkMessageQuestion_String(b *testing.B) {
	var (
		mq = MessageQuestion{
			Name: "test",
			Type: RecordTypeA,
		}

		x int
	)

	for ; x < b.N; x++ {
		_ = mq.String()
	}
}

// # 2021-11-14
//
// goos: linux
// goarch: amd64
// pkg: git.sr.ht/~shulhan/pakakeh.go/lib/dns
// cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
// BenchmarkMessageQuestion_unpack-8       35717178                35.78 ns/op            8 B/op          1 allocs/op
func BenchmarkMessageQuestion_unpack(b *testing.B) {
	var (
		packet = []byte{
			0x01, 'a',
			0x01, 'B',
			0x01, 'c',
			0x00,
			0x00, 0x01,
			0x00, 0x01,
		}
		mq = MessageQuestion{}

		x   int
		err error
	)

	for ; x < b.N; x++ {
		err = mq.unpack(packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}
