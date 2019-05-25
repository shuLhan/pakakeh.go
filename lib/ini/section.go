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
	mode      lineMode
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
		mode: lineModeSection,
		name: name,
	}

	sec.nameLower = strings.ToLower(sec.name)

	if len(subName) > 0 {
		sec.mode |= lineModeSubsection
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
// If key is empty, no variable will be appended.
// If value is empty, it will be set to true.
// If key and value already exist, no variable will be appended.
// Use set() or replaceAll() to set existing value without duplication.
//
// It will return true if new variable is appended, otherwise it will return
// false.
//
func (sec *section) add(key, value string) bool {
	if len(key) == 0 {
		return false
	}
	if len(value) == 0 {
		value = varValueTrue
	}

	keyLower := strings.ToLower(key)

	for x := 0; x < len(sec.vars); x++ {
		if !isLineModeVar(sec.vars[x].mode) {
			continue
		}
		if sec.vars[x].keyLower != keyLower {
			continue
		}
		if sec.vars[x].value == value {
			return false
		}
	}

	v := &variable{
		mode:     lineModeValue,
		key:      key,
		keyLower: keyLower,
		value:    value,
	}

	sec.vars = append(sec.vars, v)

	return true
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

	if v.mode&lineModeSingle == lineModeSingle ||
		v.mode&lineModeValue == lineModeValue ||
		v.mode&lineModeMulti == lineModeMulti {
		if len(v.value) == 0 {
			v.value = varValueTrue
		}
		v.keyLower = strings.ToLower(v.key)
	}

	sec.vars = append(sec.vars, v)
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
// getVariable return the last variable that have the same key.
// The key MUST have been converted to lowercase.
//
func (sec *section) getVariable(key string) (idx int, v *variable) {
	idx = len(sec.vars) - 1
	for ; idx >= 0; idx-- {
		if !isLineModeVar(sec.vars[idx].mode) {
			continue
		}
		if sec.vars[idx].keyLower == key {
			v = sec.vars[idx]
			return
		}
	}

	return 0, nil
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
// The key MUST be not empty and has been converted to lowercase.
// If value is empty, it will be set to true.
//
func (sec *section) set(key, value string) bool {
	if len(sec.vars) == 0 || len(key) == 0 {
		return false
	}

	key = strings.ToLower(key)

	_, v := sec.getVariable(key)
	if v == nil {
		return false
	}
	if len(value) == 0 {
		value = varValueTrue
	}

	v.value = value

	return true
}

//
// unset remove the last variable with name `key` on current section.
//
// On success, where a variable removed or one variable is removed, it will
// return true, otherwise it will be removed.
//
func (sec *section) unset(key string) bool {
	if len(key) == 0 {
		return false
	}

	key = strings.ToLower(key)

	idx, v := sec.getVariable(key)
	if v == nil {
		return false
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
