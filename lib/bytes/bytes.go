// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package bytes provide a library for working with byte or slice of bytes.
package bytes

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"

	inbytes "git.sr.ht/~shulhan/pakakeh.go/internal/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

// AppendInt16 append an int16 value into slice of byte.
func AppendInt16(data []byte, v int16) []byte {
	data = append(data, byte(v>>8))
	data = append(data, byte(v))
	return data
}

// AppendInt32 append an int32 value into slice of byte.
func AppendInt32(data []byte, v int32) []byte {
	data = append(data, byte(v>>24))
	data = append(data, byte(v>>16))
	data = append(data, byte(v>>8))
	data = append(data, byte(v))
	return data
}

// AppendInt64 append an int64 value into slice of byte.
func AppendInt64(data []byte, v int64) []byte {
	data = append(data, byte(v>>56))
	data = append(data, byte(v>>48))
	data = append(data, byte(v>>40))
	data = append(data, byte(v>>32))
	data = append(data, byte(v>>24))
	data = append(data, byte(v>>16))
	data = append(data, byte(v>>8))
	data = append(data, byte(v))
	return data
}

// AppendUint16 append an uint16 value into slice of byte.
func AppendUint16(data []byte, v uint16) []byte {
	data = append(data, byte(v>>8))
	data = append(data, byte(v))
	return data
}

// AppendUint32 append an uint32 value into slice of byte.
func AppendUint32(data []byte, v uint32) []byte {
	data = append(data, byte(v>>24))
	data = append(data, byte(v>>16))
	data = append(data, byte(v>>8))
	data = append(data, byte(v))
	return data
}

// AppendUint64 append an uint64 value into slice of byte.
func AppendUint64(data []byte, v uint64) []byte {
	data = append(data, byte(v>>56))
	data = append(data, byte(v>>48))
	data = append(data, byte(v>>40))
	data = append(data, byte(v>>32))
	data = append(data, byte(v>>24))
	data = append(data, byte(v>>16))
	data = append(data, byte(v>>8))
	data = append(data, byte(v))
	return data
}

// Concat merge one or more []byte or string in args into slice of
// byte.
// Any type that is not []byte or string in args will be ignored.
func Concat(args ...interface{}) (out []byte) {
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			out = append(out, []byte(v)...)
		case []byte:
			out = append(out, v...)
		}
	}
	return out
}

// Copy slice of bytes from parameter.
func Copy(src []byte) (dst []byte) {
	if len(src) == 0 {
		return
	}
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}

// CutUntilToken cut text until we found token.
//
// If token found, it will return all bytes before token, position of byte
// after token, and true.
//
// If no token found, it will return false.
//
// If checkEsc is true, token that is prefixed with escaped character ('\')
// will be skipped, the escape character will be removed.
func CutUntilToken(text, token []byte, startAt int, checkEsc bool) (cut []byte, pos int, found bool) {
	var isEsc bool

	textlen := len(text)
	tokenlen := len(token)
	if tokenlen == 0 {
		return text, -1, false
	}
	if startAt < 0 {
		startAt = 0
	}

	for pos = startAt; pos < textlen; pos++ {
		// Check if the escape character is used to escaped the
		// token ...
		if checkEsc && text[pos] == '\\' {
			if isEsc {
				// escaped already, its mean double '\\'
				cut = append(cut, '\\')
				isEsc = false
			} else {
				isEsc = true
			}
			continue
		}
		if text[pos] != token[0] {
			if isEsc {
				// ... turn out its not escaping token.
				cut = append(cut, '\\')
				isEsc = false
			}
			cut = append(cut, text[pos])
			continue
		}

		// We found the first token character.
		// Lets check if its match with all content of token.
		found = IsTokenAt(text, token, pos)
		if !found {
			if isEsc {
				// ... turn out its not escaping token.
				cut = append(cut, '\\')
				isEsc = false
			}
			cut = append(cut, text[pos])
			continue
		}

		// Found it, but if its prefixed with escaped char, then
		// we assumed it as non breaking token.
		if isEsc {
			cut = append(cut, token...)
			pos = pos + tokenlen - 1
			isEsc = false
			continue
		}

		// We found the token match in `text` at `p`
		return cut, pos + tokenlen, true
	}

	// We did not found it...
	return cut, pos, false
}

// EncloseRemove given a text, find the leftToken and rightToken and cut
// the content in between them and return it with status true.
// Keep doing it until no more leftToken and rightToken found.
//
// If no leftToken or rightToken is found, it will return text as is and
// false.
func EncloseRemove(text, leftToken, rightToken []byte) (cut []byte, found bool) {
	lidx := TokenFind(text, leftToken, 0)
	if lidx < 0 {
		return text, false
	}
	ridx := TokenFind(text, rightToken, lidx+1)
	if ridx < 0 {
		return text, false
	}

	cut = make([]byte, 0, len(text[:lidx])+len(text[ridx:]))
	cut = append(cut, text[:lidx]...)
	cut = append(cut, text[ridx+len(rightToken):]...)
	cut, _ = EncloseRemove(cut, leftToken, rightToken)

	return cut, true
}

// EncloseToken find "token" in "text" and enclose it with bytes from
// "leftcap" and "rightcap".
// If at least one token found, it will return modified text with true status.
// If no token is found, it will return the same text with false status.
func EncloseToken(text, token, leftcap, rightcap []byte) (
	newtext []byte,
	found bool,
) {
	enclosedLen := len(token)

	startat := 0
	for {
		foundat := TokenFind(text, token, startat)

		if foundat < 0 {
			newtext = append(newtext, text[startat:]...)
			break
		}

		newtext = append(newtext, text[startat:foundat]...)
		newtext = append(newtext, leftcap...)
		newtext = append(newtext, token...)
		newtext = append(newtext, rightcap...)

		startat = foundat + enclosedLen
	}
	if startat > 0 {
		found = true
	}

	return newtext, found
}

// InReplace replace any characters in "text" that is not in "allowed" with
// character "c".
// The replacement occur inside the "text" backing storage, which means the
// passed "text" will changes and returned.
func InReplace(text, allowed []byte, c byte) []byte {
	if len(text) == 0 {
		return nil
	}

	var found bool
	for x := 0; x < len(text); x++ {
		found = false
		for y := 0; y < len(allowed); y++ {
			if text[x] == allowed[y] {
				found = true
				break
			}
		}
		if !found {
			text[x] = c
		}
	}
	return text
}

// Indexes returns the index of the all instance of "token" in "text", or nil
// if no "token" found.
func Indexes(text, token []byte) (idxs []int) {
	if len(text) == 0 || len(token) == 0 {
		return nil
	}

	offset := 0
	for {
		idx := bytes.Index(text, token)
		if idx == -1 {
			break
		}
		idxs = append(idxs, offset+idx)
		skip := idx + len(token)
		offset += skip
		text = text[skip:]
	}
	return idxs
}

// IsTokenAt return true if `text` at index `p` match with `token`,
// otherwise it will return false.
// Empty token always return false.
func IsTokenAt(text, token []byte, p int) bool {
	textlen := len(text)
	tokenlen := len(token)
	if tokenlen == 0 {
		return false
	}
	if p < 0 {
		p = 0
	}

	if p+tokenlen > textlen {
		return false
	}

	for x := 0; x < tokenlen; x++ {
		if text[p] != token[x] {
			return false
		}
		p++
	}
	return true
}

// MergeSpaces convert sequences of white spaces into single space ' '.
func MergeSpaces(in []byte) (out []byte) {
	var isSpace bool
	for _, c := range in {
		if c == ' ' || c == '\t' || c == '\v' || c == '\f' || c == '\n' || c == '\r' {
			isSpace = true
			continue
		}
		if isSpace {
			out = append(out, ' ')
			isSpace = false
		}
		out = append(out, c)
	}
	if isSpace {
		out = append(out, ' ')
	}

	return out
}

// ParseHexDump parse the default output of [hexdump](1) utility from
// parameter in back into stream of byte.
//
// An example of default output of hexdump is
//
//	0000000 7865 5f70 6964 2f72 0000 0000 0000 0000
//	0000010 0000 0000 0000 0000 0000 0000 0000 0000
//	*
//	0000060 0000 0000 3030 3030 3537 0035 3030 3130
//
// The first column is the address and the rest of the column is the data.
//
// Each data column is 16-bit words in big-endian order, so in the above
// example, the first byte would be 65, second byte is 78 and so on.
// If parameter networkByteOrder is true, the first byte would be 78, second
// by is 65, and so on.
//
// The asterisk "*" means that the address from 0000020 to 0000050 is equal
// to the previous line, 0000010.
//
// [hexdump]: https://man.archlinux.org/man/hexdump.1
func ParseHexDump(in []byte, networkByteOrder bool) (out []byte, err error) {
	var (
		logp        = `ParseHexDump`
		parser      = NewParser(in, []byte(" \n"))
		d      byte = 255 // Just to make the first for-loop pass.

		token      []byte
		vint64     int64
		x          int
		isAsterisk bool
	)
	for d != 0 {
		// Read the address.
		token, d = parser.Read()
		if len(token) == 0 {
			break
		}
		if len(token) == 1 {
			if token[0] != '*' {
				break
			}
			isAsterisk = true
			continue
		}

		vint64, err = strconv.ParseInt(string(token), 16, 64)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		if isAsterisk {
			if len(out) > 0 {
				var start = len(out)
				if start < 16 {
					start = 0
				} else {
					start -= 16
				}
				var (
					prevRow      = out[start:]
					identicalRow = int((vint64 - int64(len(out))) / 16)
				)
				for x = 0; x < identicalRow; x++ {
					out = append(out, prevRow...)
				}
			}
		}

		// Read the two-hex, 16-bit words.
		for x = 0; x < 8; x++ {
			token, d = parser.Read()
			if len(token) == 0 {
				break
			}

			vint64, err = strconv.ParseInt(string(token), 16, 64)
			if err != nil {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}

			switch len(token) {
			case 2:
				out = append(out, byte(vint64))
			case 4:
				if networkByteOrder {
					out = append(out, byte(vint64>>8))
					out = append(out, byte(vint64))
				} else {
					out = append(out, byte(vint64))
					out = append(out, byte(vint64>>8))
				}
			}

			if d == '\n' {
				break
			}
		}
		// Ignore trailing characters.
		if d != '\n' {
			parser.SkipLine()
		}
	}
	return out, nil
}

// PrintHex will print each byte in slice as hexadecimal value into N column
// length.
func PrintHex(title string, data []byte, col int) {
	var (
		start, x int
		c        byte
	)
	fmt.Print(title)
	for x, c = range data {
		if x%col == 0 {
			if x > 0 {
				fmt.Print(" ||")
			}
			for y := start; y < x; y++ {
				if data[y] >= 33 && data[y] <= 126 {
					fmt.Printf(" %c", data[y])
				} else {
					fmt.Print(" .")
				}
			}
			fmt.Printf("\n%4d -", x)
			start = x
		}

		fmt.Printf(" %02X", c)
	}
	rest := col - (x % col)
	if rest > 0 {
		for y := 1; y < rest; y++ {
			fmt.Print("   ")
		}
		fmt.Print(" ||")
	}
	for y := start; y <= x; y++ {
		if data[y] >= 33 && data[y] <= 126 {
			fmt.Printf(" %c", data[y])
		} else {
			fmt.Print(" .")
		}
	}

	fmt.Println()
}

// ReadHexByte read two hexadecimal characters from "data" start from index
// "x" and convert them to byte.
// It will return the byte and true if its read exactly two hexadecimal
// characters, otherwise it will return 0 and false.
func ReadHexByte(data []byte, x int) (b byte, ok bool) {
	if x < 0 {
		return 0, false
	}
	if len(data) < x+2 {
		return 0, false
	}
	var y = 4
	for y >= 0 {
		switch {
		case data[x] >= '0' && data[x] <= '9':
			b |= (data[x] - '0') << y
		case data[x] >= 'A' && data[x] <= 'F':
			b |= (data[x] - ('A' - 10)) << y
		case data[x] >= 'a' && data[x] <= 'f':
			b |= (data[x] - ('a' - 10)) << y
		default:
			return 0, false
		}
		y -= 4
		x++
	}

	return b, true
}

// ReadInt16 read int16 value from "data" start at index "x".
// It will return 0 if "x" is out of range.
func ReadInt16(data []byte, x uint) (v int16) {
	if x+1 >= uint(len(data)) {
		return 0
	}
	v = int16(data[x]) << 8
	v |= int16(data[x+1])
	return v
}

// ReadInt32 read int32 value from "data" start at index "x".
// It will return 0 if "x" is out of range.
func ReadInt32(data []byte, x uint) (v int32) {
	if x+3 >= uint(len(data)) {
		return 0
	}
	v = int32(data[x]) << 24
	v |= int32(data[x+1]) << 16
	v |= int32(data[x+2]) << 8
	v |= int32(data[x+3])
	return v
}

// ReadUint16 read uint16 value from "data" start at index "x".
// If x is out of range, it will return 0.
func ReadUint16(data []byte, x uint) (v uint16) {
	if x+1 >= uint(len(data)) {
		return 0
	}
	v = uint16(data[x]) << 8
	v |= uint16(data[x+1])
	return v
}

// ReadUint32 read uint32 value from "data" start at index "x".
// If x is out of range, it will return 0.
func ReadUint32(data []byte, x uint) (v uint32) {
	if x+3 >= uint(len(data)) {
		return 0
	}
	v = uint32(data[x]) << 24
	v |= uint32(data[x+1]) << 16
	v |= uint32(data[x+2]) << 8
	v |= uint32(data[x+3])
	return v
}

// RemoveSpaces remove all spaces from input in.
func RemoveSpaces(in []byte) (out []byte) {
	var c byte
	out = make([]byte, 0, len(in))
	for _, c = range in {
		if ascii.IsSpace(c) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// SkipAfterToken skip all bytes until matched "token" is found and return the
// index after the token and boolean true.
//
// If "checkEsc" is true, token that is prefixed with escaped character
// '\' will be considered as non-match token.
//
// If no token found it will return -1 and boolean false.
func SkipAfterToken(text, token []byte, startAt int, checkEsc bool) (int, bool) {
	textlen := len(text)
	escaped := false
	if startAt < 0 {
		startAt = 0
	}

	p := startAt
	for ; p < textlen; p++ {
		// Check if the escape character is used to escaped the
		// token.
		if checkEsc && text[p] == '\\' {
			escaped = true
			continue
		}
		if text[p] != token[0] {
			if escaped {
				escaped = false
			}
			continue
		}

		// We found the first token character.
		// Lets check if its match with all content of token.
		found := IsTokenAt(text, token, p)

		// False alarm ...
		if !found {
			if escaped {
				escaped = false
			}
			continue
		}

		// Its matched, but if its prefixed with escaped char, then
		// we assumed it as non breaking token.
		if checkEsc && escaped {
			escaped = false
			continue
		}

		// We found the token at `p`
		p += len(token)
		return p, true
	}

	return -1, false
}

// SnippetByIndexes take snippet in between of each index with minimum
// snippet length.  The sniplen is the length before and after index, not the
// length of all snippet.
func SnippetByIndexes(s []byte, indexes []int, sniplen int) (snippets [][]byte) {
	var start, end int
	for _, idx := range indexes {
		start = idx - sniplen
		if start < 0 {
			start = 0
		}
		end = idx + sniplen
		if end > len(s) {
			end = len(s)
		}

		snippets = append(snippets, s[start:end])
	}

	return snippets
}

// SplitEach split the data into n number of bytes.
// If n is less or equal than zero, it will return the data as chunks.
func SplitEach(data []byte, n int) (chunks [][]byte) {
	if n <= 0 {
		chunks = append(chunks, data)
		return chunks
	}

	var (
		size  = len(data)
		rows  = (size / n)
		total int
	)
	for x := 0; x < rows; x++ {
		chunks = append(chunks, data[total:total+n])
		total += n
	}
	if total < size {
		chunks = append(chunks, data[total:])
	}
	return chunks
}

// TokenFind return the first index of matched token in text, start at custom
// index.
// If "startat" parameter is less than 0, then it will be set to 0.
// If token is empty or no token found it will return -1.
func TokenFind(text, token []byte, startat int) (at int) {
	return inbytes.TokenFind(text, token, startat)
}

// TrimNull remove 0 value ("\0" or NULL in C) at leading and trailing of
// in.
func TrimNull(in []byte) (out []byte) {
	var start int
	for ; start < len(in); start++ {
		if in[start] != 0 {
			break
		}
	}

	var end = len(in) - 1
	for ; end > start; end-- {
		if in[end] != 0 {
			break
		}
	}
	end++

	return in[start:end]
}

// WordIndexes returns the index of the all instance of word in s as long as
// word is separated by space or at the beginning or end of s.
func WordIndexes(s []byte, word []byte) (idxs []int) {
	tmp := Indexes(s, word)
	if len(tmp) == 0 {
		return nil
	}

	for _, idx := range tmp {
		x := idx - 1
		if x >= 0 {
			if !unicode.IsSpace(rune(s[x])) {
				continue
			}
		}
		x = idx + len(word)
		if x >= len(s) {
			idxs = append(idxs, idx)
			continue
		}
		if !unicode.IsSpace(rune(s[x])) {
			continue
		}
		idxs = append(idxs, idx)
	}

	return idxs
}

// DumpPrettyTable write each byte in slice data as hexadecimal, ASCII
// character, and integer with 8 columns width.
func DumpPrettyTable(w io.Writer, title string, data []byte) {
	const ncol = 8

	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprint(w, "          |  0  1  2  3  4  5  6  7 | 01234567 |   0   1   2   3   4   5   6   7 |\n")
	fmt.Fprint(w, "          |  8  9  A  B  C  D  E  F | 89ABCDEF |   8   9   A   B   C   D   E   F |\n")

	var (
		chunks = SplitEach(data, ncol)
		chunk  []byte
		x      int
		y      int
		c      byte
	)
	for x, chunk = range chunks {
		fmt.Fprintf(w, `%#08x|`, x*ncol)

		// Print as hex.
		for y, c = range chunk {
			fmt.Fprintf(w, ` %02x`, c)
		}
		for y++; y < ncol; y++ {
			fmt.Fprint(w, `   `)
		}

		// Print as char.
		fmt.Fprint(w, ` | `)
		for y, c = range chunk {
			if c >= 33 && c <= 126 {
				fmt.Fprintf(w, `%c`, c)
			} else {
				fmt.Fprint(w, `.`)
			}
		}
		for y++; y < ncol; y++ {
			fmt.Fprint(w, ` `)
		}

		// Print as integer.
		fmt.Fprint(w, ` |`)
		for y, c = range chunk {
			fmt.Fprintf(w, ` %3d`, c)
		}
		for y++; y < ncol; y++ {
			fmt.Fprint(w, `    `)
		}
		fmt.Fprintf(w, " |%d\n", x*ncol)
	}
}

// WriteUint16 write uint16 value "v" into "data" start at position "x".
// If x is out range, the data will not change.
func WriteUint16(data []byte, x uint, v uint16) {
	if x+1 >= uint(len(data)) {
		return
	}
	data[x] = byte(v >> 8)
	data[x+1] = byte(v)
}

// WriteUint32 write uint32 value into "data" start at position "x".
// If x is out range, the data will not change.
func WriteUint32(data []byte, x uint, v uint32) {
	if x+3 >= uint(len(data)) {
		return
	}
	data[x] = byte(v >> 24)
	data[x+1] = byte(v >> 16)
	data[x+2] = byte(v >> 8)
	data[x+3] = byte(v)
}
