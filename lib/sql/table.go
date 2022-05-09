// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"database/sql"
	"fmt"
	"strings"
)

// Table represent a tuple or table in database.
//
// A table has Name, PrimaryKey, and list of Row.
type Table struct {
	Name       string // Table name, required.
	PrimaryKey string // Primary key of table, optional.
	Rows       []Row  // The row or data in the table, optional.
}

// Insert all rows into table, one by one.
//
// On success, it will return list of ID, if table has primary key.
func (table *Table) Insert(tx *sql.Tx) (ids []int64, err error) {
	for _, row := range table.Rows {
		names, holders, values := row.ExtractSQLFields(DefaultPlaceHolder)
		if len(names) == 0 {
			continue
		}

		q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			table.Name, strings.Join(names, ","),
			strings.Join(holders, ","))

		res, err := tx.Exec(q, values...)
		if err != nil {
			return nil, err
		}

		id, err := res.LastInsertId()
		if err != nil {
			continue
		}

		ids = append(ids, id)
	}

	return ids, nil
}
