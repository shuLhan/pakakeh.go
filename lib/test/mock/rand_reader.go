// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

// RandReader implement [io.Reader] for mocking crypto [rand.Reader].
// To provide predictable result, the RandReader is seeded with the same
// slice of bytes.
// A call to Read will fill the passed bytes with those seed.
type RandReader struct {
	seed    []byte
	counter int
}

// NewRandReader create new random reader using seed as generator.
// The longer the seed, the longer the random values become unique.
func NewRandReader(seed []byte) (r *RandReader) {
	r = &RandReader{
		seed: seed,
	}
	return r
}

// Read fill the raw bytes with seed.
// If raw length larger than the seed, it will be filled with the same seed
// until all bytes filled.
//
// For example, given seed as "abc" (length is three), and raw length is
// five, then Read will return "abcab".
func (rr *RandReader) Read(raw []byte) (n int, err error) {
	var (
		expn = len(raw)

		nwrite int
	)

	for n < expn {
		nwrite = copy(raw[n:], rr.seed[rr.counter:])
		n += nwrite
	}

	// Increment the counter to make the seed start from next byte, so
	// the next Read will return different result but still predictable.
	rr.counter++
	if rr.counter == len(rr.seed) {
		rr.counter = 0
	}

	return n, nil
}
