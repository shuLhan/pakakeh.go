package ini

import (
	"bytes"
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
			k: bytes.TrimSpace(k),
			v: bytes.TrimSpace(v),
			c: comment,
		})
	}
}
