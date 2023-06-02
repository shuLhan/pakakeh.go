// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ascii provide a library for working with ASCII characters.
package ascii

import (
	"math/rand"

	"github.com/shuLhan/share/internal/asciiset"
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

	// capitalLetters contains list of upper case characters in ASCII.
	capitalLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// smallLetters contains list of lower case characters in ASCII.
	smallLetters = "abcdefghijklmnopqrstuvwxyz"
	// digits contains list of decimal characters.
	digits = "0123456789"
)

var (
	// Spaces contains list of white spaces in ASCII.
	Spaces = []byte{'\t', '\n', '\v', '\f', '\r', ' '}
)

var lettersSet, _ = asciiset.MakeASCIISet(Letters)
var capitalLettersSet, _ = asciiset.MakeASCIISet(capitalLetters)
var smallLettersSet, _ = asciiset.MakeASCIISet(smallLetters)
var hexaLettersSet, _ = asciiset.MakeASCIISet(HexaLetters)
var digitsSet, _ = asciiset.MakeASCIISet(digits)
var spacesSet, _ = asciiset.MakeASCIISet(string(Spaces))

// IsAlnum will return true if byte is ASCII alphanumeric character, otherwise
// it will return false.
func IsAlnum(b byte) bool {
	return IsAlpha(b) || IsDigit(b)
}

// IsAlpha will return true if byte is ASCII alphabet character, otherwise
// it will return false.
func IsAlpha(b byte) bool {
	return lettersSet.Contains(b)
}

// IsDigit will return true if byte is ASCII digit, otherwise it will return
// false.
func IsDigit(b byte) bool {
	return digitsSet.Contains(b)
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
	return hexaLettersSet.Contains(b)
}

// IsSpace will return true if byte is ASCII white spaces character,
// otherwise it will return false.
func IsSpace(b byte) bool {
	return spacesSet.Contains(b)
}

// Random generate random sequence of value from source with fixed length.
//
// This function assume that random generator has been seeded.
func Random(source []byte, n int) []byte {
	b := make([]byte, n)
	for x := 0; x < n; x++ {
		b[x] = source[rand.Intn(len(source))]
	}
	return b
}

// ToLower convert slice of ASCII characters to lower cases, in places, which
// means it will return the same slice instead of creating new one.
func ToLower(data []byte) []byte {
	for x := 0; x < len(data); x++ {
		if !capitalLettersSet.Contains(data[x]) {
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
		if !smallLettersSet.Contains(data[x]) {
			continue
		}
		data[x] -= 32
	}
	return data
}
