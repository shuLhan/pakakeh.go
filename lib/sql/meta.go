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

	listParam []*param

	// ListHolder contains list of column holder, as in "?" or "$x",
	// depends on the driver.
	ListHolder []string

	// Stores the holders for bulk insert, for example
	// "[[$1 $2] [$1 $3]]".
	bulkHolders [][]string

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

// BindName binds the column name for bulk insert.
// The argument unique=false allow setting the value only once in multiple
// insert-values.
//
// See the examples in BulkInsert_postgres and BulkInsert_nonPostgres.
func (meta *Meta) BindName(name string, unique bool) (err error) {
	logp := `BindName`

	for _, cur := range meta.listParam {
		if cur.name == name {
			return fmt.Errorf(`%s: duplicate name %s`, logp, name)
		}
	}
	param := &param{
		name:   name,
		unique: unique,
	}
	meta.listParam = append(meta.listParam, param)
	meta.ListName = append(meta.ListName, name)
	return nil
}

// BindValues bind the values for bulk insert.
// The number of arguments must match with the number of calls to
// [Meta.BindName].
//
// See the examples in BulkInsert_postgres and BulkInsert_nonPostgres.
func (meta *Meta) BindValues(values ...any) (err error) {
	logp := `BindValues`
	if len(values) != len(meta.listParam) {
		return fmt.Errorf(`%s: invalid number of arguments, want %d`,
			logp, len(meta.listParam))
	}
	holders := []string{}
	var holder string
	for x, param := range meta.listParam {
		if meta.driver == DriverNamePostgres {
			holder = meta.bindOnPostgres(param, values[x])
		} else {
			holder = meta.bindOnOthers(param, values[x])
		}
		holders = append(holders, holder)
	}
	meta.bulkHolders = append(meta.bulkHolders, holders)
	return nil
}

func (meta *Meta) bindOnPostgres(prm *param, val any) string {
	if !prm.unique && prm.holder != "" {
		return prm.holder
	}
	meta.nholder++
	meta.Index = append(meta.Index, meta.nholder)
	prm.holder = fmt.Sprintf(`$%d`, meta.nholder)
	meta.ListHolder = append(meta.ListHolder, prm.holder)
	meta.ListValue = append(meta.ListValue, val)
	return prm.holder
}

func (meta *Meta) bindOnOthers(prm *param, val any) string {
	if prm.holder == "" {
		prm.holder = DefaultPlaceHolder
	}
	meta.nholder++
	meta.Index = append(meta.Index, meta.nholder)
	meta.ListHolder = append(meta.ListHolder, prm.holder)
	meta.ListValue = append(meta.ListValue, val)
	return prm.holder
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

	if val != nil {
		meta.nholder++
		meta.Index = append(meta.Index, meta.nholder)
		if meta.driver == DriverNamePostgres {
			whereOp.holder = fmt.Sprintf(`$%d`, meta.nholder)
		} else {
			whereOp.holder = DefaultPlaceHolder
		}
	}

	meta.listOpWhere = append(meta.listOpWhere, *whereOp)

	return whereOp
}

// BulkHolders returns the string query for bulk INSERT-VALUES.
// For example, with Postgresql driver, it will return "($1,$2),($3,$4)".
func (meta *Meta) BulkHolders() string {
	var sb strings.Builder
	for x, holders := range meta.bulkHolders {
		if x > 0 {
			sb.WriteString(`,`)
		}
		sb.WriteString(`(`)
		sb.WriteString(strings.Join(holders, `,`))
		sb.WriteString(`)`)
	}
	return sb.String()
}

// Delete returns the DML for DELETE query.
func (meta *Meta) Delete(tableName string) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE %s`, tableName,
		meta.WhereFields())
}

// Holders returns string of holders joined with comma ",";
// for example "$1, $2, ...", for INSERT-VALUES.
func (meta *Meta) Holders() string {
	return strings.Join(meta.ListHolder, `,`)
}

// Insert returns the DML for INSERT query.
func (meta *Meta) Insert(tableName string) string {
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`, tableName,
		meta.Names(), meta.Holders())
}

// InsertValues returns the list of value for executing INSERT statement.
func (meta *Meta) InsertValues() []any {
	return meta.ListValue
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

// ScanHolders returns list of variables for [sql.Row.Scan] or
// [sql.Rows.Scan].
func (meta *Meta) ScanHolders() []any {
	return meta.ListValue
}

// Select returns the DML SELECT for tableName.
func (meta *Meta) Select(tableName string) (q string) {
	q = fmt.Sprintf(`SELECT %s FROM %s`, meta.Names(), tableName)
	if len(meta.listOpWhere) != 0 {
		q += fmt.Sprintf(` WHERE %s`, meta.WhereFields())
	}
	return q
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

// Update returns the DML UPDATE query for tableName.
func (meta *Meta) Update(tableName string) (q string) {
	q = fmt.Sprintf(`UPDATE %s SET %s`, tableName, meta.UpdateFields())
	if len(meta.listOpWhere) != 0 {
		q += fmt.Sprintf(` WHERE %s`, meta.WhereFields())
	}
	return q
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
		if op.val != nil {
			list = append(list, op.val)
		}
	}
	return list
}
