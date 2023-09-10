// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package text

import (
	"bytes"
	"fmt"
	"strconv"
)

// Chunk represent subset of line, contain starting position and slice of
// bytes in line.
type Chunk struct {
	V       []byte
	StartAt int
}

// JoinChunks all chunk's values using `sep` as separator and return it as
// string.
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

// MarshalJSON encode the Chunk into JSON value.
func (chunk Chunk) MarshalJSON() ([]byte, error) {
	var bb bytes.Buffer

	bb.WriteString(`{"StartAt":`)
	bb.WriteString(strconv.Itoa(chunk.StartAt))
	bb.WriteString(`,"V":`)
	bb.WriteString(strconv.Quote(string(chunk.V)))
	bb.WriteString(`}`)

	return bb.Bytes(), nil
}

func (chunk Chunk) String() string {
	return fmt.Sprintf("{StartAt:%d,V:%s}", chunk.StartAt, chunk.V)
}
