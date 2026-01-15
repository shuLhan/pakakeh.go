// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package ini

import (
	"os"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	var (
		src []byte
		err error
		x   int
	)

	src, err = os.ReadFile("testdata/input.ini")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for x = 0; x < b.N; x++ {
		_, err = Parse(src)
		if err != nil {
			b.Fatal(err)
		}
	}
}
