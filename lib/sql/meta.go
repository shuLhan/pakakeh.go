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
	// kind of DML, like SELECT or UPDATE.
	// The kind affect the [Bind] and [BindWhere].
	kind DMLKind

	driver string

	// ListName contains list of column name.
	ListName []string

	// ListHolder contains list of column holder, as in "?" or "$x",
	// depends on the driver.
	ListHolder []string

	// ListValue contains list of column values, either for insert or
	// select.
	ListValue []any

	// ListWhereCond contains list of condition to be joined with
	// ListHolder.
	// The text is a free form, does not need to be a column name.
	ListWhereCond []string

	// ListWhereValue contains list of values for where condition.
	ListWhereValue []any

	// Index collect all holder integer value, as in "1,2,3,...".
	Index []any

	nholder int
}

// NewMeta create new Meta using specific driver name.
// The driver affect the ListHolder value.
func NewMeta(driverName string, dmlKind DMLKind) (meta *Meta) {
	meta = &Meta{
		driver: driverName,
		kind:   dmlKind,
	}
	return meta
}

// Bind column name and variable for DML INSERT, SELECT, or UPDATE.
// It is a no-op for DML DELETE.
func (meta *Meta) Bind(colName string, val any) {
	if meta.kind == DMLKindDelete {
		return
	}

	var (
		name string
		x    int
	)
	for x, name = range meta.ListName {
		if name == colName {
			meta.ListValue[x] = val
			return
		}
	}

	meta.ListName = append(meta.ListName, colName)
	meta.ListValue = append(meta.ListValue, val)

	if meta.kind == DMLKindInsert || meta.kind == DMLKindUpdate {
		meta.nholder++
		meta.Index = append(meta.Index, meta.nholder)
		if meta.driver == DriverNamePostgres {
			meta.ListHolder = append(meta.ListHolder, fmt.Sprintf(`$%d`, meta.nholder))
		} else {
			meta.ListHolder = append(meta.ListHolder, DefaultPlaceHolder)
		}
	}
}

// BindWhere bind value for where condition.
//
// The cond string is optional, can be a column name with operator or any
// text like "AND col=" or "OR col=".
//
// It return the length of [Meta.ListHolder].
//
// It is a no-operation for DML INSERT.
func (meta *Meta) BindWhere(cond string, val any) int {
	if meta.kind == DMLKindInsert {
		return 0
	}

	meta.ListWhereCond = append(meta.ListWhereCond, cond)
	meta.ListWhereValue = append(meta.ListWhereValue, val)

	meta.nholder++
	meta.Index = append(meta.Index, meta.nholder)
	if meta.driver == DriverNamePostgres {
		meta.ListHolder = append(meta.ListHolder, fmt.Sprintf(`$%d`, meta.nholder))
	} else {
		meta.ListHolder = append(meta.ListHolder, DefaultPlaceHolder)
	}
	return meta.nholder
}

// Holders generate string of holder, for example "$1, $2, ...", for DML
// INSERT-VALUES.
func (meta *Meta) Holders() string {
	return strings.Join(meta.ListHolder, `,`)
}

// Names generate string of column names, for example "col1, col2, ...", for
// DML INSERT or SELECT.
//
// It will return an empty string if kind is DML UPDATE or DELETE.
func (meta *Meta) Names() string {
	if meta.kind == DMLKindUpdate || meta.kind == DMLKindDelete {
		return ``
	}
	return strings.Join(meta.ListName, `,`)
}

// Sub return the child of Meta for building subquery.
func (meta *Meta) Sub() (sub *Meta) {
	sub = &Meta{
		driver:  meta.driver,
		kind:    meta.kind,
		nholder: meta.nholder,
	}
	return sub
}

// UpdateFields generate string of "col1=<holder>, col2=<holder>, ..." for
// DML UPDATE.
//
// It will return an empty string if kind is not UPDATE.
func (meta *Meta) UpdateFields() string {
	if meta.kind != DMLKindUpdate {
		return ``
	}

	var (
		sb strings.Builder
		x  int
	)

	// Use the ListName as pivot, since ListHolder can expanded by
	// BindWhere if DML kind is UPDATE.
	for ; x < len(meta.ListName); x++ {
		if x > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `%s=%s`, meta.ListName[x], meta.ListHolder[x])
	}
	return sb.String()
}

// UpdateValues return the merged of ListValue and ListWhereValue for DML
// UPDATE.
//
// It will return nil if kind is not DML UPDATE.
func (meta *Meta) UpdateValues() (listVal []any) {
	if meta.kind != DMLKindUpdate {
		return nil
	}
	listVal = append(listVal, meta.ListValue...)
	listVal = append(listVal, meta.ListWhereValue...)
	return listVal
}

// WhereFields merge the ListWhereCond and ListHolder.
//
// It will return an empty string if kind is DML INSERT.
func (meta *Meta) WhereFields() string {
	if meta.kind == DMLKindInsert {
		return ``
	}

	var (
		off int
		sb  strings.Builder
		x   int
	)

	if meta.kind == DMLKindUpdate || meta.kind == DMLKindInsert {
		off = len(meta.ListValue)
	}
	for ; x < len(meta.ListWhereCond); x++ {
		if x > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, `%s%s`, meta.ListWhereCond[x], meta.ListHolder[off+x])
	}
	return sb.String()
}

// WhereHolders generate string of holder, for example "$1,$2, ...", based
// on number of item added with [Meta.BindWhere].
// Similar to method Holders but for where condition.
//
// It will return an empty string if kind is DML INSERT.
func (meta *Meta) WhereHolders() string {
	if meta.kind == DMLKindInsert {
		return ``
	}
	return strings.Join(meta.ListHolder, `,`)
}
