// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package text

import (
	"bytes"
	"fmt"
	"strconv"
)

// Line represent line number and slice of bytes as string.
type Line struct {
	V []byte
	N int
}

// MarshalJSON encode the Line into JSON value.
func (line Line) MarshalJSON() ([]byte, error) {
	var bb bytes.Buffer

	bb.WriteString(`{"N":`)
	bb.WriteString(strconv.Itoa(line.N))
	bb.WriteString(`,"V":`)
	bb.WriteString(strconv.Quote(string(line.V)))
	bb.WriteString(`}`)

	return bb.Bytes(), nil
}

func (line Line) String() string {
	return fmt.Sprintf("{N:%d,V:%s}", line.N, line.V)
}
