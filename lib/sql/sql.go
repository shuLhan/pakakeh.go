// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package sql is an extension to standard library "database/sql.DB" that
// provide common functionality across DBMS.
//
package sql

const (
	DriverNameMysql    = "mysql"
	DriverNamePostgres = "postgres"

	DefaultPlaceHolder = "?"
)
