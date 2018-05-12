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
	varValueTrue = "true"
)

//
// Variable define the smallest building block in INI format. It represent
// empty lines, comment, section, section with subsection, and variable.
//
type Variable struct {
	mode    varMode
	lineNum int
	format  string
	secName string
	subName string
	Key     string
	Value   string
	others  string
	_key    string
}

//
// String return formatted INI variable.
//
// nolint: gocyclo, gas
func (v *Variable) String() string {
	var buf bytes.Buffer

	switch v.mode {
	case varModeEmpty:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format)
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
			_, _ = fmt.Fprintf(&buf, v.format, v.Key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true\n", v.Key)
		}
	case varModeSingle | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.Key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = true %s\n", v.Key, v.others)
		}
	case varModeValue:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.Key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.Key, v.Value)
		}
	case varModeValue | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.Key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.Key, v.Value, v.others)
		}
	case varModeMulti:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.Key)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s\n", v.Key, v.Value)
		}
	case varModeMulti | varModeComment:
		if len(v.format) > 0 {
			_, _ = fmt.Fprintf(&buf, v.format, v.Key, v.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "%s = %s %s\n", v.Key, v.Value, v.others)
		}
	}

	return buf.String()
}
