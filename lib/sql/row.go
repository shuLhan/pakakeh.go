// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import "sort"

//
// Row represent a column-name and value in a tuple.
// The map's key is the column name in database and the map's value is
// the column's value.
// This type can be used to create dynamic insert-update fields.
//
type Row map[string]interface{}

//
// ExtractSQLFields extract the column's name, column place holder (default is
// "?"), and column values; as slices.
//
// The returned names will be sorted in ascending order.
//
func (row Row) ExtractSQLFields(placeHolder string) (names, holders []string, values []interface{}) {
	if len(row) == 0 {
		return nil, nil, nil
	}
	if len(placeHolder) == 0 {
		placeHolder = DefaultPlaceHolder
	}

	names = make([]string, 0, len(row))
	holders = make([]string, 0, len(row))
	values = make([]interface{}, 0, len(row))

	for k := range row {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, k := range names {
		holders = append(holders, placeHolder)
		values = append(values, row[k])
	}

	return names, holders, values
}
