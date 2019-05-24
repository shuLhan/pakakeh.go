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
// Section represent section header in INI file format and their variables.
//
// Remember that section's name is case insensitive. When trying to
// compare section name use the NameLower field.
//
type Section struct {
	mode      varMode
	LineNum   int
	Name      string
	Sub       string
	format    string
	NameLower string
	others    string
	Vars      []*Variable
}

//
// NewSection will create new section with `name` and optional subsection
// `subName`.
// If section `name` is empty, it will return nil.
//
func NewSection(name, subName string) (sec *Section) {
	if len(name) == 0 {
		return
	}

	sec = &Section{
		mode: varModeSection,
		Name: name,
	}

	sec.NameLower = strings.ToLower(sec.Name)

	if len(subName) > 0 {
		sec.mode |= varModeSubsection
		sec.Sub = subName
	}

	return
}

func (sec *Section) add(v *Variable) {
	if v == nil {
		return
	}

	if v.mode&varModeSingle == varModeSingle ||
		v.mode&varModeValue == varModeValue ||
		v.mode&varModeMulti == varModeMulti {
		if len(v.Value) == 0 {
			v.Value = varValueTrue
		}
		v.KeyLower = strings.ToLower(v.Key)
	}

	sec.Vars = append(sec.Vars, v)
}

//
// getFirstIndex will return the first index of variable `key`. If current
// section have duplicate `key` it will return true.
// If no variable with key found it will return -1 and false.
//
func (sec *Section) getFirstIndex(key string) (idx int, dup bool) {
	idx = -1
	n := 0
	for x := 0; x < len(sec.Vars); x++ {
		if sec.Vars[x].KeyLower != key {
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
// Set will replace variable with matching key with value.
// (1) If key is empty, no variable will be changed neither added, and it will
// return false.
// (2) If section contains two or more variable with the same `key`, it will
// return false.
// (3) If no variable key matched, the new variable will be added to list.
// (4) If value is empty, it will be set to true.
//
func (sec *Section) Set(key, value string) bool {
	// (1)
	if len(key) == 0 {
		return false
	}

	keylow := strings.ToLower(key)

	idx, dup := sec.getFirstIndex(keylow)

	// (2)
	if dup {
		return false
	}

	// (3)
	if idx < 0 {
		sec.add(&Variable{
			mode:  varModeValue,
			Key:   key,
			Value: value,
		})

		return true
	}

	// (4)
	if len(value) == 0 {
		sec.Vars[idx].Value = varValueTrue
	} else {
		sec.Vars[idx].Value = value
	}

	return true
}

//
// Add append variable with `key` and `value` to current section.
//
// If section contains the same key, the value will not be replaced.
// Use `Set` or `ReplaceAll` to set existing value with out without
// duplication.
//
// If key is empty, no variable will be appended.
// If value is empty, it will set to true.
//
func (sec *Section) Add(key, value string) {
	if len(key) == 0 {
		return
	}
	v := &Variable{
		mode:  varModeValue,
		Key:   key,
		Value: value,
	}
	sec.add(v)
}

//
// AddComment to section as variable.
//
func (sec *Section) AddComment(comment string) {
	b0 := comment[0]

	if b0 != tokHash && b0 != tokSemiColon {
		comment = "# " + comment
	}

	v := &Variable{
		mode:   varModeComment,
		format: "\t%s\n",
		others: comment,
	}

	sec.add(v)
}

//
// Unset remove the variable with name `key` on current section.
//
// (1) If key is empty, no variable will be removed, and it will return true.
// (2) If current section contains two or more variables with the same key,
// no variables will be removed and it will return false.
//
// On success, where no variable removed or one variable is removed, it will
// return true.
//
func (sec *Section) Unset(key string) bool {
	// (1)
	if len(key) == 0 {
		return true
	}

	keylow := strings.ToLower(key)

	idx, dup := sec.getFirstIndex(keylow)

	// (2)
	if dup {
		return false
	}

	if idx < 0 {
		return true
	}

	copy(sec.Vars[idx:], sec.Vars[idx+1:])
	sec.Vars[len(sec.Vars)-1] = nil
	sec.Vars = sec.Vars[:len(sec.Vars)-1]

	return true
}

//
// UnsetAll remove all variables with `key`.
//
func (sec *Section) UnsetAll(key string) {
	if len(key) == 0 {
		return
	}

	var (
		vars   []*Variable
		ok     bool
		keylow = strings.ToLower(key)
	)

	for x := 0; x < len(sec.Vars); x++ {
		if sec.Vars[x].KeyLower == keylow {
			ok = true
			sec.Vars[x] = nil
			continue
		}
		vars = append(vars, sec.Vars[x])
	}

	if ok {
		sec.Vars = vars
	}
}

//
// ReplaceAll change the value of variable reference with `key` into new
// `value`. This is basically `UnsetAll` and `Add`.
//
// If no variable found, the new variable with `key` and `value` will be
// added.
// If section contains duplicate keys, all duplicate keys will be
// removed, and replaced with one key only.
//
func (sec *Section) ReplaceAll(key, value string) {
	sec.UnsetAll(key)
	sec.Add(key, value)
}

//
// Get will return the last variable value based on key.
// If no key found it will return default value and false.
//
func (sec *Section) Get(key, def string) (val string, ok bool) {
	val = def
	if len(sec.Vars) == 0 || len(key) == 0 {
		return
	}

	x := len(sec.Vars) - 1
	key = strings.ToLower(key)

	for ; x >= 0; x-- {
		if sec.Vars[x].KeyLower != key {
			continue
		}

		val = sec.Vars[x].Value
		ok = true
		break
	}

	return
}

//
// Gets all variable values that have the same key under section from top to
// bottom.
// If no key found it will return default values and false.
//
func (sec *Section) Gets(key string, defs []string) (vals []string, ok bool) {
	if len(sec.Vars) == 0 || len(key) == 0 {
		return defs, false
	}

	key = strings.ToLower(key)
	for x := 0; x < len(sec.Vars); x++ {
		if sec.Vars[x].KeyLower == key {
			vals = append(vals, sec.Vars[x].Value)
		}
	}
	if len(vals) == 0 {
		return defs, false
	}
	return vals, true
}

//
// String return formatted INI section header.
//
func (sec *Section) String() string {
	var buf bytes.Buffer

	switch sec.mode {
	case varModeSection:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.Name)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s]\n", sec.Name)
		}
	case varModeSection | varModeComment:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.Name, sec.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s] %s\n", sec.Name, sec.others)
		}
	case varModeSection | varModeSubsection:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.Name, sec.Sub)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s \"%s\"]\n", sec.Name, sec.Sub)
		}
	case varModeSection | varModeSubsection | varModeComment:
		if len(sec.format) > 0 {
			_, _ = fmt.Fprintf(&buf, sec.format, sec.Name, sec.Sub, sec.others)
		} else {
			_, _ = fmt.Fprintf(&buf, "[%s \"%s\"] %s\n", sec.Name, sec.Sub, sec.others)
		}
	}

	return buf.String()
}
