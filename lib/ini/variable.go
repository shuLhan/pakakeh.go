// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
)

const (
	varValueTrue = "true"
)

//
// variable define the smallest building block in INI format. It represent
// empty lines, comment, section, section with subsection, and variable.
//
// Remember that variable's key is case insensitive. If you want to
// create variable, set the KeyLower to their lowercase value, and if you
// want to compare variable, use the KeyLower value.
//
type variable struct {
	mode     lineMode
	lineNum  int
	format   string
	secName  string
	subName  string
	key      string
	keyLower string
	value    string
	others   string
}

//
// String return formatted INI variable.
//
func (v *variable) String() string {
	var buf bytes.Buffer

	switch v.mode {
	case lineModeEmpty:
		if len(v.format) > 0 {
			_, _ = fmt.Fprint(&buf, v.format)
		}
	case lineModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s\n", v.others)
		}
	case lineModeSingle:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true\n", v.key)
		}
	case lineModeSingle | lineModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true %s\n", v.key, v.others)
		}
	case lineModeValue:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.value)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.key, v.value)
		}
	case lineModeValue | lineModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.value, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.key, v.value, v.others)
		}
	case lineModeMulti:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.value)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.key, v.value)
		}
	case lineModeMulti | lineModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.value, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.key, v.value, v.others)
		}
	}

	return buf.String()
}
