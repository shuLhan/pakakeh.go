// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// DecodeQP decode DKIM quoted printable text.
//
func DecodeQP(raw []byte) (out []byte) {
	if len(raw) == 0 {
		return nil
	}

	out = make([]byte, 0, len(raw))

	for x := 0; x < len(raw); x++ {
		if libbytes.IsSpace(raw[x]) {
			continue
		}
		if raw[x] == '=' {
			if x+2 < len(raw) {
				x++
				b, ok := libbytes.ReadHexByte(raw, x)
				if ok {
					out = append(out, b)
					x++
					continue
				}
				x--
			}
		}
		out = append(out, raw[x])
	}

	return out
}
