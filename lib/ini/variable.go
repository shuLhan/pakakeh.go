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

type variable struct {
	mode     varMode
	lineNum  int
	format   []byte
	secName  []byte
	subName  []byte
	key      []byte
	value    []byte
	others   []byte
	secLower []byte
	keyLower []byte
	vars     []*variable
}

//
// String return formatted INI variable.
//
// nolint: gocyclo, gas
func (v *variable) String() string {
	var buf bytes.Buffer
	format := string(v.format)

	switch v.mode {
	case varModeEmpty:
		_, _ = fmt.Fprintf(&buf, format)
	case varModeComment:
		_, _ = fmt.Fprintf(&buf, format, v.others)
	case varModeSingle:
		_, _ = fmt.Fprintf(&buf, format, v.key)
	case varModeSingle | varModeComment:
		_, _ = fmt.Fprintf(&buf, format, v.key, v.others)
	case varModeValue:
		_, _ = fmt.Fprintf(&buf, format, v.key)
	case varModeValue | varModeComment:
		_, _ = fmt.Fprintf(&buf, format, v.key, v.others)
	case varModeMulti:
		_, _ = fmt.Fprintf(&buf, format, v.key)
	case varModeMulti | varModeComment:
		_, _ = fmt.Fprintf(&buf, format, v.key, v.others)
	}

	return buf.String()
}
