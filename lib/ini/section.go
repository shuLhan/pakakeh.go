// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"fmt"
	"strings"
)

// Section represent section header in INI file format and their variables.
//
// Remember that section's name is case insensitive.
type Section struct {
	name      string
	sub       string
	format    string
	nameLower string

	vars []*variable

	mode    lineMode
	lineNum int
}

// newSection will create new section with `name` and optional subsection
// `subName`.
// If section `name` is empty, it will return nil.
func newSection(name, subName string) (sec *Section) {
	if len(name) == 0 {
		return
	}

	sec = &Section{
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

// Name return the section's name.
func (sec *Section) Name() string {
	return sec.name
}

// SubName return subsection's name.
func (sec *Section) SubName() string {
	return sec.sub
}

// String return formatted INI section header.
func (sec *Section) String() string {
	var buf bytes.Buffer

	switch sec.mode {
	case lineModeSection:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.name)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s]\n", sec.name)
		}
	case lineModeSection | lineModeSubsection:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.name, sec.sub)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s \"%s\"]\n", sec.name, sec.sub)
		}
	}

	return buf.String()
}

// Val return the last defined variable key in section.
func (sec *Section) Val(key string) string {
	val, _ := sec.get(key, "")
	return val
}

// Vals return all variables in section as slice of string.
func (sec *Section) Vals(key string) []string {
	vals, _ := sec.gets(key, nil)
	return vals
}

// add append variable with `key` and `value` to current section.
//
// If key is empty, no variable will be appended.
// If key and value already exist, no variable will be appended.
// Use set() or replaceAll() to set existing value without duplication.
//
// It will return true if new variable is appended, otherwise it will return
// false.
func (sec *Section) add(key, value string) bool {
	if len(key) == 0 {
		return false
	}

	keyLower := strings.ToLower(key)

	idx := -1
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
		idx = x
	}

	v := &variable{
		mode:     lineModeKeyValue,
		key:      key,
		keyLower: keyLower,
		value:    value,
	}

	sec.vars = append(sec.vars, v)
	if idx >= 0 {
		idx++
		copy(sec.vars[idx+1:], sec.vars[idx:])
		sec.vars[idx] = v
	}

	return true
}

// addUniqValue add a new variable with uniq value to section.
// If variable with the same key and value found, that variable will be moved
// to end of list, to make the last declared variable still at the end of
// list.
func (sec *Section) addUniqValue(key, value string) {
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
		mode:     lineModeKeyValue,
		key:      key,
		keyLower: keyLower,
		value:    value,
	}
	sec.vars = append(sec.vars, v)
}

func (sec *Section) addVariable(v *variable) {
	if v == nil {
		return
	}

	if v.mode == lineModeKeyValue {
		v.keyLower = strings.ToLower(v.key)
	}

	sec.vars = append(sec.vars, v)
}

// appendVar append the variable v after the last non-empty line.
func (sec *Section) appendVar(v *variable) {
	var x = len(sec.vars) - 1
	for ; x >= 0; x-- {
		if sec.vars[x].mode != lineModeEmpty {
			break
		}
	}
	if x < 0 {
		sec.vars = append(sec.vars, v)
		return
	}
	var tmp = make([]*variable, 0, len(sec.vars)+1)
	tmp = append(tmp, sec.vars[:x+1]...)
	tmp = append(tmp, v)
	tmp = append(tmp, sec.vars[x+1:]...)
	sec.vars = tmp
}

// get will return the last variable value based on key.
// If no key found it will return default value and false.
func (sec *Section) get(key, def string) (val string, ok bool) {
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

// getVariable return the last variable that have the same key.
// The key MUST have been converted to lowercase.
func (sec *Section) getVariable(key string) (idx int, v *variable) {
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

// gets all variable values that have the same key under section from top to
// bottom.
// If no key found it will return default values and false.
func (sec *Section) gets(key string, defs []string) (vals []string, ok bool) {
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

// merge other Section variables on this section, ignoring empty or comment
// mode.
func (sec *Section) merge(other *Section) {
	for x := 0; x < len(other.vars); x++ {
		if !isLineModeVar(other.vars[x].mode) {
			continue
		}
		sec.vars = append(sec.vars, other.vars[x])
	}
}

// replaceAll change the value of variable reference with `key` into new
// `value`. This is basically `unsetAll` and `Add`.
//
// If no variable found, the new variable with `key` and `value` will be
// added.
// If section contains duplicate keys, all duplicate keys will be
// removed, and replaced with one key only.
func (sec *Section) replaceAll(key, value string) {
	sec.unsetAll(key)
	sec.add(key, value)
}

// set will replace variable with matching key with value.
// The key MUST be not empty and has been converted to lowercase.
// If value is empty, it will be set to true.
func (sec *Section) set(key, value string) bool {
	if len(key) == 0 {
		return false
	}

	var (
		keyLower = strings.ToLower(key)

		v *variable
	)

	_, v = sec.getVariable(keyLower)
	if v == nil {
		v = &variable{
			mode:     lineModeKeyValue,
			key:      key,
			keyLower: keyLower,
			value:    value,
		}
		sec.appendVar(v)
		return true
	}

	v.value = strings.TrimSpace(value)
	if len(v.format) > 0 {
		v.rawValue = make([]byte, 0, len(value)+1)
		v.rawValue = append(v.rawValue, ' ')
		v.rawValue = append(v.rawValue, []byte(value)...)
	}

	return true
}

// unset remove the last variable with name `key` on current section.
//
// On success, where a variable removed or one variable is removed, it will
// return true, otherwise it will be removed.
func (sec *Section) unset(key string) bool {
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

// unsetAll remove all variables with `key`.
func (sec *Section) unsetAll(key string) {
	if len(key) == 0 {
		return
	}

	vars := make([]*variable, 0, len(sec.vars))
	key = strings.ToLower(key)

	for x := 0; x < len(sec.vars); x++ {
		if sec.vars[x].keyLower != key {
			// Ignore the last empty line.
			if x == len(sec.vars)-1 &&
				sec.vars[x].mode == lineModeEmpty {
				continue
			}
			vars = append(vars, sec.vars[x])
		}
	}

	sec.vars = vars
}
