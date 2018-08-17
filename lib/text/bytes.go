package text

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	errInvalidSyntax = "%s: invalid syntax at %d"
)

const (
	bDoubleQuote = '"'
	bRevSolidus  = '\\'
	bSolidus     = '/'
	bBackspace   = '\b'
	bFormFeed    = '\f'
	bLineFeed    = '\n'
	bCarReturn   = '\r'
	bTab         = '\t'
)

//
// BytesJSONEscape escape the following character: `"` (quotation mark),
// `\` (reverse solidus), `/` (solidus), `\b` (backspace), `\f` (formfeed),
// `\n` (newline), `\r` (carriage return`), `\t` (horizontal tab), and control
// character from 0 - 31.
//
// References:
//
// * https://tools.ietf.org/html/rfc7159#page-8
//
func BytesJSONEscape(in []byte) []byte {
	var buf bytes.Buffer

	for x := 0; x < len(in); x++ {
		if in[x] == bDoubleQuote || in[x] == bRevSolidus || in[x] == bSolidus {
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

//
// BytesJSONUnescape unescape JSON bytes, reversing what BytesJSONEscape do.
//
// If strict is true, any unknown control character will be returned as error.
// For example, in string "\x", "x" is not valid control character, and the
// function will return empty string and error.
// If strict is false, it will return "x".
//
func BytesJSONUnescape(in []byte, strict bool) ([]byte, error) {
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
			if in[x] == bDoubleQuote || in[x] == bRevSolidus || in[x] == bSolidus {
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
