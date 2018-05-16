package text

//
// StringJSONEscape escape the following character: `"` (quotation mark),
// `\` (reverse solidus), `/` (solidus), `\b` (backspace), `\f` (formfeed),
// `\n` (newline), `\r` (carriage return`), `\t` (horizontal tab), and control
// character from 0 - 31.
//
// References
//
// * https://tools.ietf.org/html/rfc7159#page-8
//
func StringJSONEscape(in string) string {
	if len(in) == 0 {
		return in
	}

	bin := []byte(in)
	bout := BytesJSONEscape(bin)

	return string(bout)
}
