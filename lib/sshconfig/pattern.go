// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package sshconfig

import (
	"bytes"
	"io"
	"path/filepath"
)

type pattern struct {
	value    string
	isNegate bool
}

func newPattern(s string) (pat *pattern) {
	pat = &pattern{}
	if s[0] == '!' {
		pat.isNegate = true
		pat.value = s[1:]
	} else {
		pat.value = s
	}
	return pat
}

// MarshalText encode the pattern back to ssh_config format.
func (pat *pattern) MarshalText() (text []byte, err error) {
	var buf bytes.Buffer

	if pat.isNegate {
		buf.WriteByte('!')
	}
	buf.WriteString(pat.value)

	return buf.Bytes(), nil
}

// WriteTo marshal the pattern into text and write it to w.
func (pat *pattern) WriteTo(w io.Writer) (n int64, err error) {
	var text []byte
	text, _ = pat.MarshalText()

	var c int
	c, err = w.Write(text)
	return int64(c), err
}

// isMatch will return true if input string match with regex and isNegate is
// false; otherwise it will return false.
func (pat *pattern) isMatch(s string) bool {
	ok, err := filepath.Match(pat.value, s)
	if err != nil {
		return false
	}
	if ok {
		return !pat.isNegate
	}
	return false
}
