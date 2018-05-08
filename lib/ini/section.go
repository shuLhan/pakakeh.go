package ini

import (
	"bytes"
	"fmt"
)

type sectionMode uint

const (
	sectionModeNone sectionMode = 1 << iota
	sectionModeNormal
	sectionModeSub
)

type section struct {
	m       sectionMode
	name    []byte
	subName []byte
	vars    []*variable
}

//
// pushVar will push new variable to list if no key exist or replace existing
// value if it's exist.
//
func (sec *section) pushVar(mode varMode, k, v, comment []byte) {
	switch mode {
	case varModeNewline:
		sec.vars = append(sec.vars, varNewline)

	case varModeComment:
		sec.vars = append(sec.vars, &variable{
			m: mode,
			k: nil,
			v: nil,
			c: comment,
		})

	case varModeNormal:
		sec.vars = append(sec.vars, &variable{
			m: mode,
			k: k,
			v: v,
			c: comment,
		})
	}
}

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
			fmt.Printf("sec: %s, var: %s %s\n", sec.name,
				string(sec.vars[x].k),
				string(sec.vars[x].v))
		}
		if sec.vars[x].m != varModeNormal {
			continue
		}
		if !bytes.Equal(sec.vars[x].k, key) {
			continue
		}

		val = sec.vars[x].v
		ok = true
		break
	}

	return
}
