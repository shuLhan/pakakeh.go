// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dkim

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

// DecodeQP decode DKIM quoted printable text.
func DecodeQP(raw []byte) (out []byte) {
	if len(raw) == 0 {
		return nil
	}

	out = make([]byte, 0, len(raw))

	var x int
	for ; x < len(raw); x++ {
		if ascii.IsSpace(raw[x]) {
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

// Canonicalize a simple or relaxed input of DKIM-Signature value by removing
// the value of tag "b=" and CRLF at the end.
//
// For example, "v=1; b=base64; bh=base64\r\n" would become
// "v=1; b=; bh=base64".
func Canonicalize(raw []byte) (out []byte) {
	// Find "b=" ...
	x := 0
	for ; x < len(raw); x++ {
		if raw[x] == '=' {
			if x > 1 && raw[x-1] == 'b' {
				x++
				break
			}
		}
	}
	if x == len(raw) {
		return nil
	}
	out = append(out, raw[:x]...)

	// Skip until ';' ...
	for ; x < len(raw); x++ {
		if raw[x] == ';' {
			out = append(out, raw[x:]...)
			break
		}
	}

	// Remove CRLF at the end ...
	x = len(out)
	if out[x-2] == '\r' && out[x-1] == '\n' {
		out = out[:x-2]
	}

	return out
}
