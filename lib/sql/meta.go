// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"fmt"
	"strings"
)

// Meta contains the DML meta data, including driver name, list of column
// names, list of column holders, and list of values.
type Meta struct {
	driver string

	// ListName contains list of column name.
	ListName []string

	// ListHolder contains list of column holder, as in "?" or "$x",
	// depends on driver.
	ListHolder []string

	// ListValue contains list of column values, either for insert or
	// select.
	ListValue []any

	// ListWhereHolder contains list of column holder, as in "?" or
	// "$x", depends on driver; expanded by calling AddWhere.
	ListWhereHolder []string

	// ListWhereValue contains list of values for where condition.
	ListWhereValue []any
}

// NewMeta create new Meta using specific driver name.
// The driver affect the ListHolder value.
func NewMeta(driverName string) (meta *Meta) {
	meta = &Meta{
		driver: driverName,
	}
	return meta
}

// Add column name and variable.
func (meta *Meta) Add(colName string, val any) {
	meta.ListName = append(meta.ListName, colName)

	if meta.driver == DriverNamePostgres {
		meta.ListHolder = append(meta.ListHolder, fmt.Sprintf(`$%d`, len(meta.ListName)))
	} else {
		meta.ListHolder = append(meta.ListHolder, DefaultPlaceHolder)
	}

	meta.ListValue = append(meta.ListValue, val)
}

// AddWhere add value for where condition.
// It return the length of ListWhereValue in list after addition.
func (meta *Meta) AddWhere(val any) int {
	meta.ListWhereValue = append(meta.ListWhereValue, val)

	if meta.driver == DriverNamePostgres {
		meta.ListWhereHolder = append(meta.ListWhereHolder, fmt.Sprintf(`$%d`, len(meta.ListWhereValue)))
	} else {
		meta.ListWhereHolder = append(meta.ListWhereHolder, DefaultPlaceHolder)
	}

	return len(meta.ListWhereValue)
}

// Names generate string of column names, for example "col1, col2, ...".
func (meta *Meta) Names() string {
	return strings.Join(meta.ListName, `,`)
}

// Holders generate string of holder, for example "$1, $2, ...".
func (meta *Meta) Holders() string {
	return strings.Join(meta.ListHolder, `,`)
}

// UpdateFields generate string of "col1=<holder>, col2=<holder>, ...".
func (meta *Meta) UpdateFields() string {
	var (
		sb strings.Builder
		x  int
	)
	for ; x < len(meta.ListName); x++ {
		if x > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `%s=%s`, meta.ListName[x], meta.ListHolder[x])
	}
	return sb.String()
}

// WhereHolders generate string of holder, for example "$1,$2, ...", based
// on number of item added with [Meta.AddWhere].
// Similar to method Holders but for where condition.
func (meta *Meta) WhereHolders() string {
	return strings.Join(meta.ListWhereHolder, `,`)
}
