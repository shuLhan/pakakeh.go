// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
