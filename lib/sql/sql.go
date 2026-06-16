// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

// Package sql is an extension to standard library [database/sql] that
// provides common functionalities across DBMS.
package sql

import "database/sql"

// List of known driver name for database connection.
const (
	DriverNameMysql    = "mysql"
	DriverNamePostgres = "postgres"
)

// DefaultPlaceHolder define default placeholder for DML, which is
// placeholder for MySQL.
const DefaultPlaceHolder = "?"

// BulkInsert executes bulk INSERT on table tableName using meta to generate
// query.
func BulkInsert(dbs Session, meta *Meta, tableName string) (result sql.Result, err error) {
	q := meta.BulkInsert(tableName)
	return dbs.Exec(q, meta.InsertValues()...)
}

// Delete executes the DELETE statement on table tableName using meta to
// generate query.
func Delete(dbs Session, meta *Meta, tableName string) (result sql.Result, err error) {
	q := meta.Delete(tableName)
	return dbs.Exec(q, meta.WhereValues()...)
}

// Insert executes the INSERT statement into table tableName using meta to
// generate query.
func Insert(dbs Session, meta *Meta, tableName string) (result sql.Result, err error) {
	q := meta.Insert(tableName)
	return dbs.Exec(q, meta.ListValue...)
}

// Select executes SELECT statement with [Session.QueryRow] using meta to
// generate query on table tableName.
func Select(dbs Session, meta *Meta, tableName string) (err error) {
	q := meta.Select(tableName)
	err = dbs.QueryRow(q, meta.WhereValues()...).Scan(meta.ListValue...)
	return err
}

// SelectManyFunc executes SELECT statement with [Session.Query] using meta to
// generate query on table tableName.
// The afterScan function is called after [sql.Rows.Scan] is completed in
// loop.
func SelectManyFunc(dbs Session, meta *Meta, tableName string, afterScan func()) (err error) {
	q := meta.Select(tableName)
	rows, err := dbs.Query(q, meta.WhereValues()...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(meta.ScanHolders()...)
		if err != nil {
			return err
		}
		afterScan()
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

// Update executes the UPDATE statement on table tableName using meta to
// generate query string.
func Update(dbs Session, meta *Meta, tableName string) (result sql.Result, err error) {
	q := meta.Update(tableName)
	return dbs.Exec(q, meta.UpdateValues()...)
}
