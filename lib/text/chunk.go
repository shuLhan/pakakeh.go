// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package text

import "fmt"

//
// Chunk represent subset of line, contain starting position and slice of
// bytes in line.
//
type Chunk struct {
	StartAt int
	V       []byte
}

//
// JoinChunks all chunk's values using `sep` as separator and return it as
// string.
//
func JoinChunks(chunks []Chunk, sep string) string {
	var out string

	for x := 0; x < len(chunks); x++ {
		if x > 0 {
			out += sep
		}
		out += string(chunks[x].V)
	}
	return out
}

func (c Chunk) String() string {
	return fmt.Sprintf("{StartAt:%d,V:%s}", c.StartAt, c.V)
}
