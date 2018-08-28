package net

import (
	libtext "github.com/shuLhan/share/lib/text"
)

//
// IsHostnameValid will return true if hostname is valid, otherwise it will
// return false.
// They must begin with alphanumeric character or "_" and end with an
// alphanumeric character.
// Host names may contain only alphanumeric characters, minus signs ("-"),
// underscore ("_"), and periods (".").
//
// See rfc952 and rfc1123.
//
func IsHostnameValid(hname []byte) bool {
	n := len(hname)
	if n == 0 {
		return false
	}
	if !libtext.IsAlnum(hname[0]) && hname[0] != '_' {
		return false
	}
	if !libtext.IsAlnum(hname[n-1]) {
		return false
	}
	for x := 1; x < n-1; x++ {
		if hname[x] == '-' || hname[x] == '_' || hname[x] == '.' {
			continue
		}
		if libtext.IsAlnum(hname[x]) {
			continue
		}
		return false
	}
	return true
}
