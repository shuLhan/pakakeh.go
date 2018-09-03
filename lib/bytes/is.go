// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

//
// IsAlpha will return true if byte is ASCII alphabet character, otherwise
// it will return false.
//
func IsAlpha(b byte) bool {
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
		return true
	}
	return false
}

//
// IsAlnum will return true if byte is ASCII alphanumeric character, otherwise
// it will return false.
//
func IsAlnum(b byte) bool {
	return IsAlpha(b) || IsDigit(b)
}

//
// IsDigit will return true if byte is ASCII digit, otherwise it will return
// false.
//
func IsDigit(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	return false
}

//
// IsDigits will return true if all bytes are ASCII digit, otherwise it will
// return false.
//
func IsDigits(data []byte) bool {
	for x := 0; x < len(data); x++ {
		if !IsDigit(data[x]) {
			return false
		}
	}
	return true
}

//
// IsHex will return true if byte is hexadecimal number, otherwise it will
// return false.
//
func IsHex(b byte) bool {
	if (b >= '1' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') {
		return true
	}
	return false
}

//
// IsSpace will return true if byte is ASCII white spaces character,
// otherwise it will return false.
//
func IsSpace(b byte) bool {
	if b == '\t' || b == '\n' || b == '\v' || b == '\f' || b == '\r' || b == ' ' {
		return true
	}
	return false
}
