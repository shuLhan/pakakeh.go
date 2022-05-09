// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"math"
)

// MapRowsElement represent a single mapping of string key to rows.
type MapRowsElement struct {
	Key   string
	Value Rows
}

// MapRows represent a list of mapping between string key and rows.
type MapRows []MapRowsElement

// insertRow will insert a row `v` into map using key `k`.
func (mapRows *MapRows) insertRow(k string, v *Row) {
	rows := Rows{}
	rows.PushBack(v)
	el := MapRowsElement{k, rows}
	(*mapRows) = append((*mapRows), el)
}

// AddRow will append a row `v` into map value if they key `k` exist in map,
// otherwise it will insert a new map element.
func (mapRows *MapRows) AddRow(k string, v *Row) {
	for x := range *mapRows {
		if (*mapRows)[x].Key == k {
			(*mapRows)[x].Value.PushBack(v)
			return
		}
	}
	// no key found on map
	mapRows.insertRow(k, v)
}

// GetMinority return map value which contain the minimum rows.
func (mapRows *MapRows) GetMinority() (keyMin string, valMin Rows) {
	min := math.MaxInt32

	for k := range *mapRows {
		v := (*mapRows)[k].Value
		l := len(v)
		if l < min {
			keyMin = (*mapRows)[k].Key
			valMin = v
			min = l
		}
	}
	return
}
