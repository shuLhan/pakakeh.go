// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

//
// Row represent a single row in table.
//
type Row map[string]interface{}

//
// ExtractSQLFields extract the column's name, column place holder (default is
// "?"), and column values; as slices.
//
func (row Row) ExtractSQLFields() (
	names, holders []string, values []interface{},
) {
	if len(row) == 0 {
		return nil, nil, nil
	}

	names = make([]string, 0, len(row))
	holders = make([]string, 0, len(row))
	values = make([]interface{}, 0, len(row))

	for k, v := range row {
		names = append(names, k)
		holders = append(holders, "?")
		values = append(values, v)
	}

	return names, holders, values
}
