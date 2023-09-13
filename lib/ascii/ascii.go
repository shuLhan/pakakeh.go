// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ascii provide a library for working with ASCII characters.
package ascii

import (
	"crypto/rand"
	"log"
	"math/big"
)

const (
	// Letters contains list of lower and upper case characters in ASCII.
	Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// LettersNumber contains list of lower and upper case characters in
	// ASCII along with numbers.
	LettersNumber = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

	// HexaLETTERS contains list of hexadecimal characters in upper cases.
	HexaLETTERS = "0123456789ABCDEF"
	// HexaLetters contains list of hexadecimal characters in lower and
	// upper cases.
	HexaLetters = "0123456789abcedfABCDEF"
	// Hexaletters contains list of hexadecimal characters in lower cases.
	Hexaletters = "0123456789abcedf"
)

var (
	// Spaces contains list of white spaces in ASCII.
	Spaces = []byte{'\t', '\n', '\v', '\f', '\r', ' '}
)

// IsAlnum will return true if byte is ASCII alphanumeric character, otherwise
// it will return false.
func IsAlnum(b byte) bool {
	return IsAlpha(b) || IsDigit(b)
}

// IsAlpha will return true if byte is ASCII alphabet character, otherwise
// it will return false.
func IsAlpha(b byte) bool {
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
		return true
	}
	return false
}

// IsDigit will return true if byte is ASCII digit, otherwise it will return
// false.
func IsDigit(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	return false
}

// IsDigits will return true if all bytes are ASCII digit, otherwise it will
// return false.
func IsDigits(data []byte) bool {
	for x := 0; x < len(data); x++ {
		if !IsDigit(data[x]) {
			return false
		}
	}
	return true
}

// IsHex will return true if byte is hexadecimal number, otherwise it will
// return false.
func IsHex(b byte) bool {
	if (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') {
		return true
	}
	return false
}

// IsSpace will return true if byte is ASCII white spaces character,
// otherwise it will return false.
func IsSpace(b byte) bool {
	if b == '\t' || b == '\n' || b == '\v' || b == '\f' || b == '\r' || b == ' ' {
		return true
	}
	return false
}

// Random generate random sequence of value from source with fixed length.
func Random(source []byte, n int) []byte {
	var (
		b   = make([]byte, n)
		max = big.NewInt(int64(len(source)))

		randv *big.Int
		err   error
		x     int
	)
	for ; x < n; x++ {
		randv, err = rand.Int(rand.Reader, max)
		if err != nil {
			log.Panicf(`Random: %s`, err)
		}

		b[x] = source[int(randv.Int64())]
	}
	return b
}

// ToLower convert slice of ASCII characters to lower cases, in places, which
// means it will return the same slice instead of creating new one.
func ToLower(data []byte) []byte {
	for x := 0; x < len(data); x++ {
		if data[x] < 'A' || data[x] > 'Z' {
			continue
		}
		data[x] += 32
	}
	return data
}

// ToUpper convert slice of ASCII characters to upper cases, in places, which
// means it will return the same slice instead of creating new one.
func ToUpper(data []byte) []byte {
	for x := 0; x < len(data); x++ {
		if data[x] < 'a' || data[x] > 'z' {
			continue
		}
		data[x] -= 32
	}
	return data
}
