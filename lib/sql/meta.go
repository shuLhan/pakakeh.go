// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package sql

import (
	"fmt"
	"strconv"
	"strings"
)

// Meta contains the DML meta data, including driver name, list of column
// names, list of column holders, and list of values.
type Meta struct {
	driver string

	// ListName contains list of column name.
	ListName []string

	listParam []*param

	// Stores the holders for bulk insert, for example
	// "[[$1 $2] [$1 $3]]".
	bulkHolders [][]string

	// ListValue contains list of column values, either for insert or
	// select.
	ListValue []any

	listJoin []opJoin

	// listOpWhere contains list of WHERE conditions.
	listOpWhere []OpWhere

	listOrderBy []opOrderBy

	nholder int

	limit int
}

// NewMeta creates new Meta using specific driver name and DML operations.
func NewMeta(driverName string) (meta *Meta) {
	meta = &Meta{
		driver: driverName,
	}
	return meta
}

// Bind binds the column name and value for DML INSERT or UPDATE.
func (meta *Meta) Bind(colName string, val any) {
	meta.bind(colName, val, true)
}

// BindJoin adds the JOIN statement for SELECT query.
func (meta *Meta) BindJoin(onTable, alias, onColumn, refColumn string) {
	op := opJoin{
		mode:      `JOIN`,
		onTable:   onTable,
		alias:     alias,
		onColumn:  onColumn,
		refColumn: refColumn,
	}
	meta.listJoin = append(meta.listJoin, op)
}

// BindLeftJoin adds the LEFT JOIN statement for SELECT query.
func (meta *Meta) BindLeftJoin(onTable, alias, onColumn, refColumn string) {
	op := opJoin{
		mode:      `LEFT JOIN`,
		onTable:   onTable,
		alias:     alias,
		onColumn:  onColumn,
		refColumn: refColumn,
	}
	meta.listJoin = append(meta.listJoin, op)
}

// BindLimit sets the limit of rows in select query.
func (meta *Meta) BindLimit(n int) {
	meta.limit = n
}

// BindOrderBy adds the column name and their order for SELECT ... ORDER BY
// query.
func (meta *Meta) BindOrderBy(name, order string) {
	op := opOrderBy{
		name:  name,
		order: order,
	}
	meta.listOrderBy = append(meta.listOrderBy, op)
}

// BindSelect binds the column name and scan holder value for DML SELECT.
func (meta *Meta) BindSelect(colName string, val any) {
	meta.bind(colName, val, false)
}

func (meta *Meta) bind(colName string, val any, isInsertUpdate bool) {
	for x, name := range meta.ListName {
		if name == colName {
			meta.ListValue[x] = val
			return
		}
	}

	meta.ListName = append(meta.ListName, colName)
	meta.ListValue = append(meta.ListValue, val)

	if isInsertUpdate {
		meta.nholder++
	}
}

// BulkBindName binds the column name for bulk insert.
// The argument unique=false allow setting the value only once in multiple
// insert-values.
//
// See the examples in [Meta.BulkInsert].
func (meta *Meta) BulkBindName(name string, unique bool) (err error) {
	logp := `BulkBindName`

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

// BulkBindValues bind the values for bulk insert.
// The number of arguments must match with the number of calls to
// [Meta.BindName].
//
// See the examples in [Meta.BulkInsert].
func (meta *Meta) BulkBindValues(values ...any) (err error) {
	logp := `BulkBindValues`
	if len(values) != len(meta.listParam) {
		return fmt.Errorf(`%s: invalid number of arguments, want %d`,
			logp, len(meta.listParam))
	}
	holders := []string{}
	var holder string
	for x, param := range meta.listParam {
		if meta.driver == DriverNamePostgres {
			holder = meta.bulkBindOnPostgres(param, values[x])
		} else {
			holder = meta.bulkBindOnOthers(param, values[x])
		}
		holders = append(holders, holder)
	}
	meta.bulkHolders = append(meta.bulkHolders, holders)
	return nil
}

func (meta *Meta) bulkBindOnPostgres(prm *param, val any) string {
	if !prm.unique && prm.holder != "" {
		return prm.holder
	}
	meta.nholder++
	prm.holder = fmt.Sprintf(`$%d`, meta.nholder)
	meta.ListValue = append(meta.ListValue, val)
	return prm.holder
}

func (meta *Meta) bulkBindOnOthers(prm *param, val any) string {
	if prm.holder == "" {
		prm.holder = DefaultPlaceHolder
	}
	meta.nholder++
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
func (meta *Meta) BindWhere(logic, col, comp string, val any) (whereOp *OpWhere) {
	whereOp = &OpWhere{
		logic:  logic,
		column: col,
		comp:   comp,
		val:    val,
	}

	if val != nil {
		meta.nholder++
		if meta.driver == DriverNamePostgres {
			whereOp.holder = fmt.Sprintf(`$%d`, meta.nholder)
		} else {
			whereOp.holder = DefaultPlaceHolder
		}
	}

	meta.listOpWhere = append(meta.listOpWhere, *whereOp)

	return whereOp
}

// BulkHolders returns the string holders for bulk VALUES in INSERT query.
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

// BulkInsert returns the string query for bulk INSERT.
// For example, with Postgresql driver, it will return
// "INSERT INTO t (x,y) VALUES ($1,$2),($3,$4)".
func (meta *Meta) BulkInsert(tableName string) string {
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s;`, tableName,
		meta.Names(), meta.BulkHolders())
}

// Delete returns the DML for DELETE query.
func (meta *Meta) Delete(tableName string) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE %s`, tableName,
		meta.WhereFields())
}

// Holders returns string of holders joined with comma ",";
// for example "$1, $2, ...", for INSERT-VALUES.
func (meta *Meta) Holders() string {
	var sb strings.Builder
	for x := range meta.ListName {
		if x > 0 {
			sb.WriteString(`,`)
		}
		if meta.driver == DriverNamePostgres {
			sb.WriteString(`$`)
			sb.WriteString(strconv.Itoa(x + 1))
		} else {
			sb.WriteString(DefaultPlaceHolder)
		}
	}
	return sb.String()
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
func (meta *Meta) Names() string {
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
	if len(meta.listJoin) != 0 {
		for _, op := range meta.listJoin {
			q += op.String()
		}
	}
	if len(meta.listOpWhere) != 0 {
		q += fmt.Sprintf(` WHERE %s`, meta.WhereFields())
	}
	if len(meta.listOrderBy) != 0 {
		q += ` ORDER BY `
		for x, op := range meta.listOrderBy {
			if x > 0 {
				q += `, `
			}
			q += op.String()
		}
	}
	if meta.limit != 0 {
		q += ` LIMIT ` + strconv.Itoa(meta.limit)
	}
	return q
}

// Sub returns the child of Meta for building subquery.
func (meta *Meta) Sub() (sub *Meta) {
	sub = &Meta{
		driver:  meta.driver,
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
func (meta *Meta) UpdateFields() string {
	var sb strings.Builder

	for x := 0; x < len(meta.ListName); x++ {
		if x > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(meta.ListName[x])
		sb.WriteString(`=`)
		if meta.driver == DriverNamePostgres {
			sb.WriteString(`$`)
			sb.WriteString(strconv.Itoa(x + 1))
		} else {
			sb.WriteString(DefaultPlaceHolder)
		}
	}
	return sb.String()
}

// UpdateValues returns the merged of ListValue and list WhereValues for DML
// UPDATE.
func (meta *Meta) UpdateValues() (listVal []any) {
	listVal = append(listVal, meta.ListValue...)
	listVal = append(listVal, meta.WhereValues()...)
	return listVal
}

// WhereFields returns the DML for WHERE query, as in "col=$? AND y!=$?...".
func (meta *Meta) WhereFields() string {
	var sb strings.Builder
	for _, op := range meta.listOpWhere {
		sb.WriteString(op.String())
	}
	return sb.String()
}

// WhereHolders returns string of holders joining by comma, for example
// "$1,$2, ...", based on number of item added with [Meta.BindWhere].
// Similar to method Holders but for where condition.
func (meta *Meta) WhereHolders() string {
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
