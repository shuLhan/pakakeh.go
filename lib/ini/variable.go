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
	rawValue []byte

	mode    lineMode
	lineNum int
}

// String return formatted INI variable.
func (v *variable) String() string {
	var (
		buf bytes.Buffer
	)

	switch v.mode {
	case lineModeEmpty:
		if len(v.format) > 0 {
			_, _ = fmt.Fprint(&buf, v.format)
		}
	case lineModeComment:
		buf.WriteString(v.format)

	case lineModeKeyOnly:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key)
		} else {
			buf.WriteString(v.key)
			buf.WriteByte('\n')
		}

	case lineModeKeyValue:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.rawValue)
		} else {
			buf.WriteString(v.key)
			buf.WriteString(" =")
			if len(v.value) > 0 {
				buf.WriteByte(' ')
				buf.WriteString(v.value)
			}
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}
