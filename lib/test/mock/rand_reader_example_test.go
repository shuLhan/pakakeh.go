// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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

		n   int
		err error
	)

	rand.Reader = rr

	for range len(seed) + 1 {
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
