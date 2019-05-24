// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
	"strings"
)

//
// section represent section header in INI file format and their variables.
//
// Remember that section's name is case insensitive. When trying to
// compare section name use the NameLower field.
//
type section struct {
	mode      varMode
	lineNum   int
	name      string
	sub       string
	format    string
	nameLower string
	others    string
	vars      []*variable
}

//
// newSection will create new section with `name` and optional subsection
// `subName`.
// If section `name` is empty, it will return nil.
//
func newSection(name, subName string) (sec *section) {
	if len(name) == 0 {
		return
	}

	sec = &section{
		mode: varModeSection,
		name: name,
	}

	sec.nameLower = strings.ToLower(sec.name)

	if len(subName) > 0 {
		sec.mode |= varModeSubsection
		sec.sub = subName
	}

	return
}

//
// String return formatted INI section header.
//
func (sec *section) String() string {
	var buf bytes.Buffer

	switch sec.mode {
	case lineModeSection:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.name)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s]\n", sec.name)
		}
	case lineModeSection | lineModeComment:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.name, sec.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s] %s\n", sec.name, sec.others)
		}
	case lineModeSection | lineModeSubsection:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.name, sec.sub)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s \"%s\"]\n", sec.name, sec.sub)
		}
	case lineModeSection | lineModeSubsection | lineModeComment:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.name, sec.sub, sec.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s \"%s\"] %s\n", sec.name, sec.sub, sec.others)
		}
	}

	return buf.String()
}

//
// add append variable with `key` and `value` to current section.
//
// If section already contains the same key, the value will not be replaced.
// Use set() or ReplaceAll() to set existing value without duplication.
// If key is empty, no variable will be appended.
// If value is empty, it will be set to true.
//
func (sec *section) add(key, value string) {
	if len(key) == 0 {
		return
	}
	v := &variable{
		mode:  lineModeValue,
		key:   key,
		value: value,
	}
	sec.addVariable(v)
}

//
// addUniqValue add a new variable with uniq value to section.
// If variable with the same key and value found, that variable will be moved
// to end of list, to make the last declared variable still at the end of
// list.
//
func (sec *section) addUniqValue(key, value string) {
	keyLower := strings.ToLower(key)
	for x := 0; x < len(sec.vars); x++ {
		if sec.vars[x].keyLower == keyLower {
			if sec.vars[x].value == value {
				tmp := sec.vars[x]
				sec.vars = append(sec.vars[:x], sec.vars[x+1:]...)
				sec.vars = append(sec.vars, tmp)
				return
			}
		}
	}
	v := &variable{
		mode:     lineModeValue,
		key:      key,
		keyLower: keyLower,
		value:    value,
	}
	sec.vars = append(sec.vars, v)
}

func (sec *section) addVariable(v *variable) {
	if v == nil {
		return
	}

	if v.mode&varModeSingle == varModeSingle ||
		v.mode&varModeValue == varModeValue ||
		v.mode&varModeMulti == varModeMulti {
		if len(v.value) == 0 {
			v.value = varValueTrue
		}
		v.keyLower = strings.ToLower(v.key)
	}

	sec.vars = append(sec.vars, v)
}

//
// getFirstIndex will return the first index of variable `key`. If current
// section have duplicate `key` it will return true.
// If no variable with key found it will return -1 and false.
//
func (sec *section) getFirstIndex(key string) (idx int, dup bool) {
	idx = -1
	n := 0
	for x := 0; x < len(sec.vars); x++ {
		if sec.vars[x].keyLower != key {
			continue
		}
		if idx < 0 {
			idx = x
		}
		n++
		if n > 1 {
			dup = true
			return
		}
	}

	return
}

//
// get will return the last variable value based on key.
// If no key found it will return default value and false.
//
func (sec *section) get(key, def string) (val string, ok bool) {
	val = def
	if len(sec.vars) == 0 || len(key) == 0 {
		return
	}

	x := len(sec.vars) - 1
	key = strings.ToLower(key)

	for ; x >= 0; x-- {
		if sec.vars[x].keyLower != key {
			continue
		}

		val = sec.vars[x].value
		ok = true
		break
	}

	return
}

//
// gets all variable values that have the same key under section from top to
// bottom.
// If no key found it will return default values and false.
//
func (sec *section) gets(key string, defs []string) (vals []string, ok bool) {
	if len(sec.vars) == 0 || len(key) == 0 {
		return defs, false
	}

	key = strings.ToLower(key)
	for x := 0; x < len(sec.vars); x++ {
		if sec.vars[x].keyLower == key {
			vals = append(vals, sec.vars[x].value)
		}
	}
	if len(vals) == 0 {
		return defs, false
	}
	return vals, true
}

//
// replaceAll change the value of variable reference with `key` into new
// `value`. This is basically `unsetAll` and `Add`.
//
// If no variable found, the new variable with `key` and `value` will be
// added.
// If section contains duplicate keys, all duplicate keys will be
// removed, and replaced with one key only.
//
func (sec *section) replaceAll(key, value string) {
	sec.unsetAll(key)
	sec.add(key, value)
}

//
// set will replace variable with matching key with value.
// If key is empty, no variable will be changed or added, and it will
// return false.
// If section contains two or more variable with the same `key`, it will
// return false.
// If no variable key matched, the new variable will be added to list.
// If value is empty, it will be set to true.
//
func (sec *section) set(key, value string) bool {
	if len(key) == 0 {
		return false
	}

	keyLower := strings.ToLower(key)

	idx, dup := sec.getFirstIndex(keyLower)
	if dup {
		return false
	}

	if idx < 0 {
		sec.addVariable(&variable{
			mode:  varModeValue,
			key:   key,
			value: value,
		})
		return true
	}

	if len(value) == 0 {
		sec.vars[idx].value = varValueTrue
	} else {
		sec.vars[idx].value = value
	}

	return true
}

//
// unset remove the variable with name `key` on current section.
//
// If key is empty, no variable will be removed, and it will return true.
//
// If current section contains two or more variables with the same key,
// no variables will be removed and it will return false.
//
// On success, where no variable removed or one variable is removed, it will
// return true.
//
func (sec *section) unset(key string) bool {
	if len(key) == 0 {
		return true
	}

	key = strings.ToLower(key)

	idx, dup := sec.getFirstIndex(key)
	if dup {
		return false
	}
	if idx < 0 {
		return true
	}

	copy(sec.vars[idx:], sec.vars[idx+1:])
	sec.vars[len(sec.vars)-1] = nil
	sec.vars = sec.vars[:len(sec.vars)-1]

	return true
}

//
// unsetAll remove all variables with `key`.
//
func (sec *section) unsetAll(key string) {
	if len(key) == 0 {
		return
	}

	var (
		vars []*variable
		ok   bool
	)
	key = strings.ToLower(key)

	for x := 0; x < len(sec.vars); x++ {
		if sec.vars[x].keyLower == key {
			ok = true
			sec.vars[x] = nil
			continue
		}
		vars = append(vars, sec.vars[x])
	}

	if ok {
		sec.vars = vars
	}
}
