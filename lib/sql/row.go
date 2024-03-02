// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"fmt"
	"sort"
)

// Row represent a column-name and value in a tuple.
// The map's key is the column name in database and the map's value is
// the column's value.
// This type can be used to create dynamic insert-update fields.
//
// Deprecated: use [Meta] instead.
type Row map[string]interface{}

// Meta convert the Row into Meta.
func (row Row) Meta(driverName string) (meta *Meta) {
	meta = &Meta{}

	if len(row) == 0 {
		return meta
	}

	meta.ListName = make([]string, 0, len(row))
	meta.ListHolder = make([]string, 0, len(row))
	meta.ListValue = make([]interface{}, 0, len(row))

	var colName string
	for colName = range row {
		meta.ListName = append(meta.ListName, colName)
	}
	sort.Strings(meta.ListName)

	var x int
	for x, colName = range meta.ListName {
		if driverName == DriverNamePostgres {
			meta.ListHolder = append(meta.ListHolder, fmt.Sprintf(`$%d`, x+1))
		} else {
			meta.ListHolder = append(meta.ListHolder, DefaultPlaceHolder)
		}
		meta.ListValue = append(meta.ListValue, row[colName])
	}
	return meta
}

// ExtractSQLFields extract the column's name, column place holder, and column
// values as slices.
//
// The driverName define the returned place holders.
// If the driverName is "postgres" then the list of holders will be returned
// as counter, for example "$1", "$2" and so on.
// If the driverName is "mysql" or empty or unknown the the list of holders
// will be returned as list of "?".
//
// The returned names will be sorted in ascending order.
func (row Row) ExtractSQLFields(driverName string) (names, holders []string, values []interface{}) {
	if len(row) == 0 {
		return nil, nil, nil
	}

	var meta = row.Meta(driverName)

	return meta.ListName, meta.ListHolder, meta.ListValue
}
