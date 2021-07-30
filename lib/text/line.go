// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package text

import "fmt"

//
// Line represent line number and slice of bytes as string.
//
type Line struct {
	N int
	V []byte
}

func (l Line) String() string {
	return fmt.Sprintf("{N:%d,V:%s}", l.N, l.V)
}
