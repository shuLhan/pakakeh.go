// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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

	for x := range len(chunks) {
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
