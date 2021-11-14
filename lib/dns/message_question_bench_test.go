// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import "testing"

/**
== 2021-11-15

goos: linux
goarch: amd64
pkg: github.com/shuLhan/share/lib/dns
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkMessageQuestion_String-8        7138899               168.3 ns/op            56 B/op          3 allocs/op
**/
func BenchmarkMessageQuestion_String(b *testing.B) {
	mq := MessageQuestion{
		Name: "test",
		Type: RecordTypeA,
	}
	for x := 0; x < b.N; x++ {
		_ = mq.String()
	}
}

/**
== 2021-11-14

goos: linux
goarch: amd64
pkg: github.com/shuLhan/share/lib/dns
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkMessageQuestion_unpack-8       35717178                35.78 ns/op            8 B/op          1 allocs/op
**/
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
		err error
	)
	mq := MessageQuestion{}
	for x := 0; x < b.N; x++ {
		err = mq.unpack(packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}
