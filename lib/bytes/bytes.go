// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bytes provide a library for working with byte or slice of bytes.
package bytes

import (
	"fmt"
)

const (
	// ASCIILetters contains list of lower and upper case characters in
	// ASCII.
	ASCIILetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// HexaLETTERS contains list of hexadecimal characters in upper cases.
	HexaLETTERS = "0123456789ABCDEF"
	// HexaLetters contains list of hexadecimal characters in lower and
	// upper cases.
	HexaLetters = "0123456789abcedfABCDEF"
	// HexaLetters contains list of hexadecimal characters in lower cases.
	Hexaletters = "0123456789abcedf"
)

var (
	// ASCIISpaces contains list of white spaces in ASCII.
	ASCIISpaces = []byte{'\t', '\n', '\v', '\f', '\r', ' '}
)

//
// AppendInt16 into slice of byte.
//
func AppendInt16(data *[]byte, v int16) {
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// AppendInt32 into slice of byte.
//
func AppendInt32(data *[]byte, v int32) {
	*data = append(*data, byte(v>>24))
	*data = append(*data, byte(v>>16))
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// AppendUint16 into slice of byte.
//
func AppendUint16(data *[]byte, v uint16) {
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// AppendUint32 into slice of byte.
//
func AppendUint32(data *[]byte, v uint32) {
	*data = append(*data, byte(v>>24))
	*data = append(*data, byte(v>>16))
	*data = append(*data, byte(v>>8))
	*data = append(*data, byte(v))
}

//
// CutUntilToken cut line until we found token.
//
// If token found, it will return all cutted bytes before token, positition of
// byte after token, and boolean true.
//
// If no token found, it will return false.
//
// If `checkEsc` is true, token that is prefixed with escaped character
// '\' will be skipped.
//
//
func CutUntilToken(line, token []byte, startAt int, checkEsc bool) ([]byte, int, bool) {
	var (
		v              []byte
		p              int
		found, escaped bool
	)

	linelen := len(line)
	tokenlen := len(token)
	if tokenlen == 0 {
		return line, -1, false
	}
	if startAt < 0 {
		startAt = 0
	}

	for p = startAt; p < linelen; p++ {
		// Check if the escape character is used to escaped the
		// token ...
		if checkEsc && line[p] == '\\' {
			if escaped {
				// escaped already, its mean double '\\'
				v = append(v, '\\')
				escaped = false
			} else {
				escaped = true
			}
			continue
		}
		if line[p] != token[0] {
			if escaped {
				// ... turn out its not escaping token.
				v = append(v, '\\')
				escaped = false
			}
			v = append(v, line[p])
			continue
		}

		// We found the first token character.
		// Lets check if its match with all content of token.
		found = IsTokenAt(line, token, p)

		// False alarm ...
		if !found {
			if escaped {
				// ... turn out its not escaping token.
				v = append(v, '\\')
				escaped = false
			}
			v = append(v, line[p])
			continue
		}

		// Found it, but if its prefixed with escaped char, then
		// we assumed it as non breaking token.
		if escaped {
			v = append(v, token...)
			p = p + tokenlen - 1
			escaped = false
			continue
		}

		// We found the token match in `line` at `p`
		return v, p + tokenlen, true
	}

	// We did not found it...
	return v, p, false
}

//
// EncloseRemove given a line, remove all bytes inside it, starting from
// `leftcap` until the `rightcap` and return cutted line and status to true.
//
// If no `leftcap` or `rightcap` is found, it will return line as is, and
// status will be false.
//
func EncloseRemove(line, leftcap, rightcap []byte) ([]byte, bool) {
	lidx := TokenFind(line, leftcap, 0)
	ridx := TokenFind(line, rightcap, lidx+1)

	if lidx < 0 || ridx < 0 || lidx >= ridx {
		return line, false
	}

	var newline []byte
	newline = append(newline, line[:lidx]...)
	newline = append(newline, line[ridx+len(rightcap):]...)
	newline, _ = EncloseRemove(newline, leftcap, rightcap)

	return newline, true
}

//
// EncloseToken will find `token` in `line` and enclose it with bytes from
// `leftcap` and `rightcap`.
// If at least one token found, it will return modified line with true status.
// If no token is found, it will return the same line with false status.
//
func EncloseToken(line, token, leftcap, rightcap []byte) (
	newline []byte,
	status bool,
) {
	enclosedLen := len(token)

	startat := 0
	for {
		foundat := TokenFind(line, token, startat)

		if foundat < 0 {
			newline = append(newline, line[startat:]...)
			break
		}

		newline = append(newline, line[startat:foundat]...)
		newline = append(newline, leftcap...)
		newline = append(newline, token...)
		newline = append(newline, rightcap...)

		startat = foundat + enclosedLen
	}
	if startat > 0 {
		status = true
	}

	return
}

//
// IsTokenAt return true if `line` at index `p` match with `token`,
// otherwise it will return false.
// Empty token always return false.
//
func IsTokenAt(line, token []byte, p int) bool {
	linelen := len(line)
	tokenlen := len(token)
	if tokenlen == 0 {
		return false
	}
	if p < 0 {
		p = 0
	}

	if p+tokenlen > linelen {
		return false
	}

	for x := 0; x < tokenlen; x++ {
		if line[p] != token[x] {
			return false
		}
		p++
	}
	return true
}

//
// PrintHex will print each byte in slice as hexadecimal value into N column
// length.
//
func PrintHex(title string, data []byte, col int) {
	fmt.Print(title)
	for x := 0; x < len(data); x++ {
		if x%col == 0 {
			fmt.Printf("\n%4d -", x)
		}

		fmt.Printf(" %02X", data[x])
	}
	fmt.Println()
}

//
// ReadInt16 will convert two bytes from data start at `x` into int16 and
// return it.
//
func ReadInt16(data []byte, x uint) int16 {
	return int16(data[x])<<8 | int16(data[x+1])
}

//
// ReadInt32 will convert four bytes from data start at `x` into int32 and
// return it.
//
func ReadInt32(data []byte, x uint) int32 {
	return int32(data[x])<<24 | int32(data[x+1])<<16 | int32(data[x+2])<<8 | int32(data[x+3])
}

//
// ReadUint16 will convert two bytes from data start at `x` into uint16 and
// return it.
//
func ReadUint16(data []byte, x uint) uint16 {
	return uint16(data[x])<<8 | uint16(data[x+1])
}

//
// ReadUint32 will convert four bytes from data start at `x` into uint32 and
// return it.
//
func ReadUint32(data []byte, x uint) uint32 {
	return uint32(data[x])<<24 | uint32(data[x+1])<<16 | uint32(data[x+2])<<8 | uint32(data[x+3])
}

//
// SkipAfterToken skip all bytes until matched token is found and return the
// index after the token and boolean true.
//
// If `checkEsc` is true, token that is prefixed with escaped character
// '\' will be considered as non-match token.
//
// If no token found it will return -1 and boolean false.
//
func SkipAfterToken(line, token []byte, startAt int, checkEsc bool) (int, bool) {
	linelen := len(line)
	escaped := false
	if startAt < 0 {
		startAt = 0
	}

	p := startAt
	for ; p < linelen; p++ {
		// Check if the escape character is used to escaped the
		// token.
		if checkEsc && line[p] == '\\' {
			escaped = true
			continue
		}
		if line[p] != token[0] {
			if escaped {
				escaped = false
			}
			continue
		}

		// We found the first token character.
		// Lets check if its match with all content of token.
		found := IsTokenAt(line, token, p)

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
		p = p + len(token)
		return p, true
	}

	return p, false
}

//
// ToLower convert slice of bytes to lower cases, in places.
//
func ToLower(data *[]byte) {
	for x := 0; x < len(*data); x++ {
		if (*data)[x] < 'A' || (*data)[x] > 'Z' {
			continue
		}
		(*data)[x] = (*data)[x] + 32
	}
}

//
// ToUpper convert slice of bytes to upper cases, in places.
//
func ToUpper(data *[]byte) {
	for x := 0; x < len(*data); x++ {
		if (*data)[x] < 'a' || (*data)[x] > 'z' {
			continue
		}
		(*data)[x] = (*data)[x] - 32
	}
}

//
// TokenFind return the first index of matched token in line, start at custom
// index.
// If "startat" parameter is less than 0, then it will be set to 0.
// If token is empty or no token found it will return -1.
//
func TokenFind(line, token []byte, startat int) (at int) {
	linelen := len(line)
	tokenlen := len(token)
	if tokenlen == 0 {
		return -1
	}
	if startat < 0 {
		startat = 0
	}

	y := 0
	at = -1
	for x := startat; x < linelen; x++ {
		if line[x] == token[y] {
			if y == 0 {
				at = x
			}
			y++
			if y == tokenlen {
				// we found it!
				return
			}
		} else {
			if at != -1 {
				// reset back
				y = 0
				at = -1
			}
		}
	}
	// x run out before y
	if y < tokenlen {
		at = -1
	}
	return
}

//
// WriteUint16 into slice of byte.
//
func WriteUint16(data *[]byte, x uint, v uint16) {
	(*data)[x] = byte(v >> 8)
	(*data)[x+1] = byte(v)
}

//
// WriteUint32 into slice of byte.
//
func WriteUint32(data *[]byte, x uint, v uint32) {
	(*data)[x] = byte(v >> 24)
	(*data)[x+1] = byte(v >> 16)
	(*data)[x+2] = byte(v >> 8)
	(*data)[x+3] = byte(v)
}
