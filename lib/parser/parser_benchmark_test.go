// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import "testing"

// Output:
//
// BenchmarkParser_Token-4          59117898                20.2 ns/op             0 B/op          0 allocs/op
//
func BenchmarkParser_Token(b *testing.B) {
	content := "abc;def"
	delims := " /;"

	p := New(content, delims)

	for x := 0; x < b.N; x++ {
		p.Token()
		p.Load(content, delims)
	}
}
