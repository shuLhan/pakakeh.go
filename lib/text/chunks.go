// Copyright 2018 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package text

// Chunks represent a set of chunk.
type Chunks []Chunk

// Join all chunk's values using `sep` as separator and return it.
func (chunks *Chunks) Join(sep string) (s string) {
	chunkslen := len(*chunks) - 1

	for x, c := range *chunks {
		s += string(c.V)

		if x < chunkslen {
			s += sep
		}
	}
	return
}
