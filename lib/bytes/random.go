// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

import (
	"math/rand"
)

//
// Random generate random sequence of value from seed with fixed length.
//
// This function assume that random generator has been seeded.
//
func Random(seed []byte, n int) []byte {
	b := make([]byte, n)
	for x := 0; x < n; x++ {
		b[x] = seed[rand.Intn(len(seed))]
	}
	return b
}
