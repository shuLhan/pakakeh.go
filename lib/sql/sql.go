// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sql is an extension to standard library "database/sql.DB" that
// provide common functionality across DBMS.
package sql

// List of known driver name for database connection.
const (
	DriverNameMysql    = "mysql"
	DriverNamePostgres = "postgres"
)

// DefaultPlaceHolder define default placeholder for DML, which is
// placeholder for MySQL.
const DefaultPlaceHolder = "?"

// JoinValues join list of slice of values into single slice.
func JoinValues(s ...[]any) (all []any) {
	var sub []any
	for _, sub = range s {
		all = append(all, sub...)
	}
	return all
}
