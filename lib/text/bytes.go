package text

import (
	"bytes"
	"fmt"
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
	bUnicode     = 'u'
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
func BytesJSONEscape(in []byte) (out []byte) {
	var (
		buf bytes.Buffer
	)

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
		if in[x] >= 0 && in[x] <= 31 {
			buf.WriteString(fmt.Sprintf("\\u%04X", in[x]))
			continue
		}

		buf.WriteByte(in[x])
	}

	return buf.Bytes()
}
