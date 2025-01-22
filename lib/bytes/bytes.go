// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package bytes provide a library for working with byte or slice of bytes.
package bytes

import (
	"bytes"
	"unicode"

	inbytes "git.sr.ht/~shulhan/pakakeh.go/internal/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

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
	for x := range len(text) {
		found = false
		for y := range len(allowed) {
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

	for x := range tokenlen {
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
	for range rows {
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
