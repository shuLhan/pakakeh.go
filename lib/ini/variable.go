// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
)

// variable define the smallest building block in INI format.
// It represent empty line, comment, section, section with subsection, and
// variable.
//
// Remember that variable's key is case insensitive, so to compare variable,
// use the KeyLower value.
type variable struct {
	format   string
	secName  string
	subName  string
	key      string
	keyLower string
	value    string

	mode    lineMode
	lineNum int

	isQuoted bool
}

// String return formatted INI variable.
func (v *variable) String() string {
	var (
		buf bytes.Buffer
		val string
	)

	if v.isQuoted {
		val = escape(v.value)
	} else {
		val = v.value
	}

	switch v.mode {
	case lineModeEmpty:
		if len(v.format) > 0 {
			_, _ = fmt.Fprint(&buf, v.format)
		}
	case lineModeComment:
		buf.WriteString(v.format)

	case lineModeValue:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, val)
		} else {
			buf.WriteString(v.key + " =")
			if len(val) > 0 {
				buf.WriteString(" " + val)
			}
			buf.WriteByte('\n')
		}
	case lineModeMulti:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, val)
		} else {
			buf.WriteString(v.key + " =")
			if len(val) > 0 {
				buf.WriteString(" " + val)
			}
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

func escape(value string) (out string) {
	var buf bytes.Buffer

	buf.Grow(len(value) + 2)

	buf.WriteByte('"')

	for _, c := range value {
		switch c {
		case '\b':
			buf.WriteString(`\b`)
		case '\n':
			buf.WriteString(`\n`)
		case '\t':
			buf.WriteString(`\t`)
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		default:
			buf.WriteRune(c)
		}
	}

	buf.WriteByte('"')

	return buf.String()
}
