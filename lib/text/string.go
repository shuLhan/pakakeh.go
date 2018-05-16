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

//
// StringJSONUnescape unescape JSON string, reversing what StringJSONEscape
// do.
//
// If strict is true, any unknown control character will be returned as error.
// For example, in string "\x", "x" is not valid control character, and the
// function will return empty string and error.
// If strict is false, it will return "x".
//
func StringJSONUnescape(in string, strict bool) (string, error) {
	if len(in) == 0 {
		return in, nil
	}

	bin := []byte(in)
	bout, err := BytesJSONUnescape(bin, strict)

	out := string(bout)

	return out, err
}
