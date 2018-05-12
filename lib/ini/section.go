// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
)

//
// Section is an alias of Variable, which represent section header in INI file
// format.
//
type Section = Variable

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
		mode:    varModeSection,
		secName: []byte(name),
	}

	sec._sec = bytes.ToLower(sec.secName)

	if len(subName) > 0 {
		sec.mode |= varModeSubsection
		sec.subName = []byte(subName)
	}

	return
}

func (sec *Section) add(v *Variable) {
	if v == nil {
		return
	}

	if len(v.value) == 0 || v.value == nil {
		v.value = varValueTrue
	}

	v._key = bytes.ToLower(v.key)

	sec.vars = append(sec.vars, v)
}

//
// getFirstIndex will return the first index of variable `key`. If current
// section have duplicate `key` it will return true.
// If no variable with key found it will return -1 and false.
//
func (sec *Section) getFirstIndex(key []byte) (idx int, dup bool) {
	idx = -1
	n := 0
	for x := 0; x < len(sec.vars); x++ {
		if !bytes.Equal(sec.vars[x]._key, key) {
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

	bkey := []byte(key)
	lowerkey := bytes.ToLower(bkey)

	idx, dup := sec.getFirstIndex(lowerkey)

	// (2)
	if dup {
		return false
	}

	// (3)
	if idx < 0 {
		sec.add(&Variable{
			mode:  varModeValue,
			key:   bkey,
			value: []byte(value),
		})

		return true
	}

	// (4)
	if len(value) == 0 {
		sec.vars[idx].value = varValueTrue
	} else {
		sec.vars[idx].value = []byte(value)
	}

	return true
}

//
// Add append variable with `key` and `value` to current section.
//
// WARNING: if section contains the same key, the value will not be replaced.
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
		key:   []byte(key),
		value: []byte(value),
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

	bkey := bytes.ToLower([]byte(key))

	idx, dup := sec.getFirstIndex(bkey)

	// (2)
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
// UnsetAll remove all variables with `key`.
//
func (sec *Section) UnsetAll(key string) {
	if len(key) == 0 {
		return
	}

	var (
		vars []*Variable
		ok   bool
	)
	bkey := bytes.ToLower([]byte(key))

	for x := 0; x < len(sec.vars); x++ {
		if bytes.Equal(sec.vars[x]._key, bkey) {
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
// If no key found it will return nil and false.
//
func (sec *Section) Get(key []byte) (val []byte, ok bool) {
	if len(sec.vars) == 0 || len(key) == 0 {
		return
	}
	x := len(sec.vars) - 1
	key = bytes.ToLower(key)

	for ; x >= 0; x-- {
		if !bytes.Equal(sec.vars[x]._key, key) {
			continue
		}

		val = sec.vars[x].value
		ok = true
		break
	}

	return
}
