// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

package net

import (
	"net"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

// IsHostnameValid will return true if hostname is valid, otherwise it will
// return false.
// They must begin with alphanumeric character or "_" and end with an
// alphanumeric character.
// Host names may contain only alphanumeric characters, minus signs ("-"),
// underscore ("_"), and periods (".").
//
// If isFQDN is true, the hostname must at least contains two labels;
// otherwise it will be invalid.
//
// See rfc952 and rfc1123.
func IsHostnameValid(hname []byte, isFQDN bool) bool {
	n := len(hname)
	if n == 0 {
		return false
	}
	if !ascii.IsAlnum(hname[0]) && hname[0] != '_' {
		return false
	}
	if !ascii.IsAlnum(hname[n-1]) {
		return false
	}
	var ndot int
	for x := 1; x < n-1; x++ {
		if hname[x] == '.' {
			ndot++
			continue
		}
		if hname[x] == '-' || hname[x] == '_' || ascii.IsAlnum(hname[x]) {
			continue
		}
		return false
	}
	if isFQDN && ndot == 0 {
		return false
	}
	return true
}

// IsIPv4 will return true if string representation of IP contains three dots,
// for example "127.0.0.1".
func IsIPv4(ip net.IP) bool {
	if ip == nil {
		return false
	}
	sip := ip.String()
	if len(sip) == 0 {
		return false
	}
	if strings.Count(sip, ".") == 3 {
		return true
	}
	return false
}

// IsIPv6 will return true if string representation of IP contains two or more
// colons ":", for example, "::1".
func IsIPv6(ip net.IP) bool {
	if ip == nil {
		return false
	}
	sip := ip.String()
	if len(sip) == 0 {
		return false
	}
	if strings.Count(sip, ":") >= 2 {
		return true
	}
	return false
}
