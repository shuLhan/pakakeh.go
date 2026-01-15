// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

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
