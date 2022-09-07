// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package text

import (
	"bytes"
	"fmt"
	"strconv"
)

// Line represent line number and slice of bytes as string.
type Line struct {
	N int
	V []byte
}

func (line Line) MarshalJSON() ([]byte, error) {
	var bb bytes.Buffer

	bb.WriteString(`{"N":`)
	bb.WriteString(strconv.Itoa(line.N))
	bb.WriteString(`,"V":`)
	bb.WriteString(strconv.Quote(string(line.V)))
	bb.WriteString(`}`)

	return bb.Bytes(), nil
}

func (l Line) String() string {
	return fmt.Sprintf("{N:%d,V:%s}", l.N, l.V)
}
