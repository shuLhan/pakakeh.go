// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql_test

import (
	"fmt"

	"github.com/shuLhan/share/lib/sql"
)

func ExampleMeta_Add_mysql() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNameMysql)
		t    = Table{}
	)

	meta.Add(`id`, &t.ID)
	meta.Add(`name`, &t.Name)

	fmt.Println(meta.Names())
	fmt.Println(meta.Holders())
	fmt.Println(meta.UpdateFields())
	// Output:
	// id,name
	// ?,?
	// id=?,name=?
}

func ExampleMeta_Add_postgres() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
		t    = Table{}
	)

	meta.Add(`id`, &t.ID)
	meta.Add(`name`, &t.Name)

	fmt.Println(meta.Names())
	fmt.Println(meta.Holders())
	fmt.Println(meta.UpdateFields())
	// Output:
	// id,name
	// $1,$2
	// id=$1,name=$2
}

func ExampleMeta_AddWhere() {
	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
		vals = []any{
			int(1000),
			string(`name`),
		}
		idx int
	)

	idx = meta.AddWhere(vals[0])
	fmt.Println(idx, meta.ListWhereValue)

	idx = meta.AddWhere(vals[1])
	fmt.Println(idx, meta.ListWhereValue)

	// Output:
	// 1 [1000]
	// 2 [1000 name]
}

func ExampleMeta_WhereHolders() {
	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
		vals = []any{
			int(1000),
			string(`name`),
		}
		v any
	)

	for _, v = range vals {
		meta.AddWhere(v)
	}
	fmt.Println(meta.WhereHolders())

	// Output:
	// $1,$2
}
