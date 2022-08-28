// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package json provide a library for working with JSON.
//
// This is an extension to standard "encoding/json" package.
package json

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	errInvalidSyntax = "%s: invalid syntax at %d"
)

const (
	bDoubleQuote = '"'
	bRevSolidus  = '\\'
	bBackspace   = '\b'
	bFormFeed    = '\f'
	bLineFeed    = '\n'
	bCarReturn   = '\r'
	bTab         = '\t'
)

// Escape the following character: `"` (quotation mark),
// `\` (reverse solidus), `\b` (backspace), `\f` (formfeed),
// `\n` (newline), `\r` (carriage return`), `\t` (horizontal tab), and control
// character from 0 - 31.
//
// References:
//
// * https://tools.ietf.org/html/rfc7159#page-8
func Escape(in []byte) []byte {
	var buf bytes.Buffer

	for x := 0; x < len(in); x++ {
		if in[x] == bDoubleQuote || in[x] == bRevSolidus {
			buf.WriteByte(bRevSolidus)
			buf.WriteByte(in[x])
			continue
		}
		if in[x] == bBackspace {
			buf.WriteByte(bRevSolidus)
			buf.WriteByte('b')
			continue
		}
		if in[x] == bFormFeed {
			buf.WriteByte(bRevSolidus)
			buf.WriteByte('f')
			continue
		}
		if in[x] == bLineFeed {
			buf.WriteByte(bRevSolidus)
			buf.WriteByte('n')
			continue
		}
		if in[x] == bCarReturn {
			buf.WriteByte(bRevSolidus)
			buf.WriteByte('r')
			continue
		}
		if in[x] == bTab {
			buf.WriteByte(bRevSolidus)
			buf.WriteByte('t')
			continue
		}
		if in[x] <= 31 {
			buf.WriteString(fmt.Sprintf("\\u%04X", in[x]))
			continue
		}

		buf.WriteByte(in[x])
	}

	return buf.Bytes()
}

// EscapeString escape the following character: `"` (quotation mark),
// `\` (reverse solidus), `\b` (backspace), `\f` (formfeed),
// `\n` (newline), `\r` (carriage return`), `\t` (horizontal tab), and control
// character from 0 - 31.
//
// # References
//
// * https://tools.ietf.org/html/rfc7159#page-8
func EscapeString(in string) string {
	if len(in) == 0 {
		return in
	}

	inb := []byte(in)
	outb := Escape(inb)

	return string(outb)
}

// ToMapStringFloat64 convert the map of string-interface{} into map of
// string-float64.
// This function convert the map's key to lower-cases and ignore zero value in
// interface{}.
// The interface{} value only accept basic numeric types and slice of byte.
func ToMapStringFloat64(in map[string]interface{}) (out map[string]float64, err error) {
	out = make(map[string]float64, len(in))

	for k, v := range in {
		var (
			f64 float64
			err error
		)

		switch vv := v.(type) {
		case string:
			f64, err = strconv.ParseFloat(vv, 64)
		case []byte:
			f64, err = strconv.ParseFloat(string(vv), 64)
		case byte:
			f64 = float64(vv)
		case float32:
			f64 = float64(vv)
		case float64:
			f64 = vv
		case int8:
			f64 = float64(vv)
		case int16:
			f64 = float64(vv)
		case int32:
			f64 = float64(vv)
		case int:
			f64 = float64(vv)
		case int64:
			f64 = float64(vv)
		case uint16:
			f64 = float64(vv)
		case uint32:
			f64 = float64(vv)
		case uint64:
			f64 = float64(vv)
		}
		if err != nil {
			return nil, err
		}
		if f64 == 0 {
			continue
		}

		k = strings.ToLower(k)
		out[k] = f64
	}
	return out, nil
}

// Unescape JSON bytes, reversing what Escape function do.
//
// If strict is true, any unknown control character will be returned as error.
// For example, in string "\x", "x" is not valid control character, and the
// function will return empty string and error.
// If strict is false, it will return "x".
func Unescape(in []byte, strict bool) ([]byte, error) {
	var (
		buf bytes.Buffer
		uni bytes.Buffer
		esc bool
	)

	for x := 0; x < len(in); x++ {
		if esc {
			if in[x] == 'u' {
				uni.Reset()
				x++

				for y := 0; y < 4 && x < len(in); x++ {
					uni.WriteByte(in[x])
					y++
				}

				dec, err := strconv.ParseUint(uni.String(), 16, 32)
				if err != nil {
					return nil, err
				}

				if dec <= 31 {
					buf.WriteByte(byte(dec))
				} else {
					buf.WriteRune(rune(dec))
				}

				esc = false
				x--
				continue
			}
			if in[x] == 't' {
				buf.WriteByte(bTab)
				esc = false
				continue
			}
			if in[x] == 'r' {
				buf.WriteByte(bCarReturn)
				esc = false
				continue
			}
			if in[x] == 'n' {
				buf.WriteByte(bLineFeed)
				esc = false
				continue
			}
			if in[x] == 'f' {
				buf.WriteByte(bFormFeed)
				esc = false
				continue
			}
			if in[x] == 'b' {
				buf.WriteByte(bBackspace)
				esc = false
				continue
			}
			if in[x] == bDoubleQuote || in[x] == bRevSolidus {
				buf.WriteByte(in[x])
				esc = false
				continue
			}

			if strict {
				err := fmt.Errorf(errInvalidSyntax, "BytesJSONUnescape", x)
				return nil, err
			}

			buf.WriteByte(in[x])
			esc = false
			continue
		}
		if in[x] == bRevSolidus {
			esc = true
			continue
		}
		buf.WriteByte(in[x])
	}

	return buf.Bytes(), nil
}

// UnescapeString unescape JSON string, reversing what EscapeString do.
//
// If strict is true, any unknown control character will be returned as error.
// For example, in string "\x", "x" is not valid control character, and the
// function will return empty string and error.
// If strict is false, it will return "x".
func UnescapeString(in string, strict bool) (string, error) {
	if len(in) == 0 {
		return in, nil
	}

	inb := []byte(in)
	outb, err := Unescape(inb, strict)

	out := string(outb)

	return out, err
}
