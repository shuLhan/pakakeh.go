// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock_test

import (
	"crypto/rand"
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test/mock"
)

func ExampleRandReader() {
	var (
		seed = []byte(`123`)
		rr   = mock.NewRandReader(seed)
		b    = make([]byte, 8)

		x   int
		n   int
		err error
	)

	rand.Reader = rr

	for x = 0; x <= len(seed); x++ {
		n, err = rand.Read(b)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(n, string(b))
	}

	// Output:
	// 8 12312312
	// 8 23232323
	// 8 33333333
	// 8 12312312
}
