package text

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
