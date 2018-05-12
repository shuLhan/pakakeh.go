// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
)

type varMode uint

const varModeEmpty varMode = 0

const (
	varModeComment    varMode = 1 << iota
	varModeSection            // 2
	varModeSubsection         // 4
	varModeSingle             // 8
	varModeValue              // 16
	varModeMulti              // 32
)

var (
	varValueTrue = []byte("true")
)

//
// Variable define the smallest building block in INI format. It represent
// empty lines, comment, section, section with subsection, and variable.
//
type Variable struct {
	mode    varMode
	lineNum int
	format  []byte
	secName []byte
	subName []byte
	key     []byte
	value   []byte
	others  []byte
	_sec    []byte
	_key    []byte
	vars    []*Variable
}

//
// String return formatted INI variable.
//
// nolint: gocyclo, gas
func (v *Variable) String() string {
	var buf bytes.Buffer
	format := string(v.format)

	switch v.mode {
	case varModeEmpty:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format)
		}
	case varModeComment:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s\n", v.others)
		}
	case varModeSection:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.secName)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s]\n", v.secName)
		}
	case varModeSection | varModeComment:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.secName, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s] %s\n", v.secName, v.others)
		}
	case varModeSection | varModeSubsection:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.secName, v.subName)
		} else {
			_, _ = fmt.Fprintf(&buf, `[%s "%s"]\n`, v.secName, v.subName)
		}
	case varModeSection | varModeSubsection | varModeComment:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.secName, v.subName, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, `[%s "%s"] %s\n`, v.secName, v.subName, v.others)
		}
	case varModeSingle:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true\n", v.key)
		}
	case varModeSingle | varModeComment:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true %s\n", v.key, v.others)
		}
	case varModeValue:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.key, v.value)
		}
	case varModeValue | varModeComment:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.key, v.value, v.others)
		}
	case varModeMulti:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.key, v.value)
		}
	case varModeMulti | varModeComment:
		if len(format) > 0 {
			_, _ = fmt.Fprintf(&buf, format, v.key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.key, v.value, v.others)
		}
	}

	return buf.String()
}
