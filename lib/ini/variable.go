// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
)

type varMode uint

const (
	varModeEmpty      varMode = 0
	varModeComment            = 1
	varModeSection            = 2
	varModeSubsection         = 4
	varModeSingle             = 8
	varModeValue              = 16
	varModeMulti              = 32
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
	mode     varMode
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
	case varModeEmpty:
		if len(v.format) > 0 {
			_, _ = fmt.Fprint(&buf, v.format)
		}
	case varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s\n", v.others)
		}
	case varModeSection:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.secName)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s]\n", v.secName)
		}
	case varModeSection | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.secName, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s] %s\n", v.secName, v.others)
		}
	case varModeSection | varModeSubsection:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.secName, v.subName)
		} else {
			_, _ = fmt.Fprintf(&buf, `[%s "%s"]\n`, v.secName, v.subName)
		}
	case varModeSection | varModeSubsection | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.secName, v.subName, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, `[%s "%s"] %s\n`, v.secName, v.subName, v.others)
		}
	case varModeSingle:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true\n", v.key)
		}
	case varModeSingle | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true %s\n", v.key, v.others)
		}
	case varModeValue:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.key, v.value)
		}
	case varModeValue | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.key, v.value, v.others)
		}
	case varModeMulti:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.key, v.value)
		}
	case varModeMulti | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.key, v.value, v.others)
		}
	}

	return buf.String()
}
