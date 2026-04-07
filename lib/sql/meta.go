// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

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

	// listOpWhere contains list of WHERE conditions.
	listOpWhere []OpWhere

	// Index collect all holder integer value, as in "1,2,3,...".
	Index []any

	nholder int
}

// NewMeta creates new Meta using specific driver name and DML operations.
// The driver affect the [Meta.ListHolder] value.
func NewMeta(driverName string, dmlKind DMLKind) (meta *Meta) {
	meta = &Meta{
		driver: driverName,
		kind:   dmlKind,
	}
	return meta
}

// Bind binds column name and variable for DML INSERT, SELECT, or UPDATE.
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

// BindWhere binds the value val for DML WHERE condition.
//
// The logic parameter defines the logical operator like "OR"
// and "AND" for joining multiple BindWhere.
//
// The col parameter defines the column name to be compared with val.
//
// The comp parameter defines the comparison for col and val, like "=" for
// equality or "!=" for non-equality.
//
// It is a no-operation for DML INSERT.
func (meta *Meta) BindWhere(logic, col, comp string, val any) (whereOp *OpWhere) {
	if meta.kind == DMLKindInsert {
		return nil
	}

	whereOp = &OpWhere{
		logic:  logic,
		column: col,
		comp:   comp,
		val:    val,
	}

	meta.nholder++
	meta.Index = append(meta.Index, meta.nholder)
	if meta.driver == DriverNamePostgres {
		whereOp.holder = fmt.Sprintf(`$%d`, meta.nholder)
	} else {
		whereOp.holder = DefaultPlaceHolder
	}

	meta.listOpWhere = append(meta.listOpWhere, *whereOp)

	return whereOp
}

// Holders returns string of holders joined with comma ",";
// for example "$1, $2, ...", for INSERT-VALUES.
func (meta *Meta) Holders() string {
	return strings.Join(meta.ListHolder, `,`)
}

// Names returns string of column names, for example "col1, col2, ...", for
// DML INSERT or SELECT.
//
// It will return an empty string if kind is UPDATE or DELETE.
func (meta *Meta) Names() string {
	if meta.kind == DMLKindUpdate || meta.kind == DMLKindDelete {
		return ``
	}
	return strings.Join(meta.ListName, `,`)
}

// Sub returns the child of Meta for building subquery.
func (meta *Meta) Sub() (sub *Meta) {
	sub = &Meta{
		driver:  meta.driver,
		kind:    meta.kind,
		nholder: meta.nholder,
	}
	return sub
}

// UpdateFields returns string of "col1=<holder>, col2=<holder>, ..." for
// DML UPDATE.
//
// It will return an empty string if kind is not DML UPDATE.
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

// UpdateValues returns the merged of ListValue and list WhereValues for DML
// UPDATE.
//
// It will return nil if kind is not DML UPDATE.
func (meta *Meta) UpdateValues() (listVal []any) {
	if meta.kind != DMLKindUpdate {
		return nil
	}
	listVal = append(listVal, meta.ListValue...)
	listVal = append(listVal, meta.WhereValues()...)
	return listVal
}

// WhereFields returns the DML for WHERE query, as in "col=$? AND y!=$?...".
// It will return an empty string if kind is DML INSERT.
func (meta *Meta) WhereFields() string {
	if meta.kind == DMLKindInsert {
		return ``
	}

	var sb strings.Builder
	for _, op := range meta.listOpWhere {
		sb.WriteString(op.String())
	}
	return sb.String()
}

// WhereHolders returns string of holders joining by comma, for example
// "$1,$2, ...", based on number of item added with [Meta.BindWhere].
// Similar to method Holders but for where condition.
//
// It will return an empty string if kind is DML INSERT.
func (meta *Meta) WhereHolders() string {
	if meta.kind == DMLKindInsert {
		return ``
	}
	var sb strings.Builder
	for x, op := range meta.listOpWhere {
		if x > 0 {
			sb.WriteString(`,`)
		}
		sb.WriteString(op.holder)
	}
	return sb.String()
}

// WhereValues returns list of values in WHERE conditions.
func (meta *Meta) WhereValues() (list []any) {
	list = make([]any, 0, len(meta.listOpWhere))
	for _, op := range meta.listOpWhere {
		list = append(list, op.val)
	}
	return list
}
