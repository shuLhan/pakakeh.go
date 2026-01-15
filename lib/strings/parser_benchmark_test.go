// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package strings

import "testing"

// Output:
//
// BenchmarkParser_Read-4          59117898                20.2 ns/op             0 B/op          0 allocs/op
func BenchmarkParser_Read(b *testing.B) {
	content := `abc;def`
	delims := ` /;`

	p := NewParser(content, delims)

	for x := 0; x < b.N; x++ {
		p.Read()
		p.Load(content, delims)
	}
}
