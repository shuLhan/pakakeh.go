// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

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
