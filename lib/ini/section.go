package ini

import (
	"bytes"
	"fmt"
)

type section variable

//
// Get will return the last key's value.
// If no key found it will return nil and false.
//
func (sec *section) Get(key []byte) (val []byte, ok bool) {
	if len(sec.vars) == 0 || len(key) == 0 {
		return
	}
	x := len(sec.vars) - 1
	key = bytes.ToLower(key)

	for ; x >= 0; x-- {
		if debug >= debugL2 {
			fmt.Printf("sec: %s, var: %s %s\n", sec.secName,
				string(sec.vars[x].key),
				string(sec.vars[x].value))
		}
		if !bytes.Equal(sec.vars[x].keyLower, key) {
			continue
		}

		val = sec.vars[x].value
		ok = true
		break
	}

	return
}

func (sec *section) addVariable(v *variable) {
	if v == nil {
		return
	}

	v.keyLower = bytes.ToLower(v.key)
	sec.vars = append(sec.vars, v)
}

//
// String return formatted INI section header.
// nolint: gas
func (sec *section) String() string {
	var buf bytes.Buffer
	format := string(sec.format)

	switch sec.mode {
	case varModeSection:
		_, _ = fmt.Fprintf(&buf, format, sec.secName)
	case varModeSection | varModeComment:
		_, _ = fmt.Fprintf(&buf, format, sec.secName, sec.others)
	case varModeSection | varModeSubsection:
		_, _ = fmt.Fprintf(&buf, format, sec.secName, sec.subName)
	case varModeSection | varModeSubsection | varModeComment:
		_, _ = fmt.Fprintf(&buf, format, sec.secName, sec.subName, sec.others)
	}

	return buf.String()
}
