// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"database/sql"
	"fmt"
)

//
// Client provide a wrapper for generic database instance.
//
type Client struct {
	DB         *sql.DB
	DriverName string
	TableNames []string // List of tables in database.
}

//
// New wrap a database client to provide additional methods.
//
func New(driverName string, db *sql.DB) (cl *Client, err error) {
	cl = &Client{
		DB:         db,
		DriverName: driverName,
	}

	return cl, nil
}

//
// FetchTableNames return the table names in current database schema sorted
// in ascending order.
//
func (cl *Client) FetchTableNames() (tableNames []string, err error) {
	var q, v string

	switch cl.DriverName {
	case DriverNameMysql, DriverNamePostgres:
		q = `
			SELECT
				table_name
			FROM
				information_schema.tables
			ORDER BY
				table_name
		`
	}

	rows, err := cl.DB.Query(q)
	if err != nil {
		return nil, fmt.Errorf("FetchTableNames: " + err.Error())
	}

	if len(cl.TableNames) > 0 {
		cl.TableNames = cl.TableNames[:0]
	}

	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			_ = rows.Close()
			return cl.TableNames, err
		}

		cl.TableNames = append(cl.TableNames, v)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return cl.TableNames, nil
}

//
// TruncateTable truncate all data on table `tableName`.
//
func (cl *Client) TruncateTable(tableName string) (err error) {
	q := `TRUNCATE TABLE ` + tableName
	_, err = cl.DB.Exec(q)
	if err != nil {
		return fmt.Errorf("TruncateTable %q: %s", tableName, err)
	}
	return nil
}
