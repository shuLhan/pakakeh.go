// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql_test

import (
	"fmt"
	"strings"

	"github.com/shuLhan/share/lib/sql"
)

func ExampleMeta_Bind() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNameMysql, sql.DMLKindSelect)
		t    = Table{}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	var q = fmt.Sprintf(`SELECT %s FROM t;`, meta.Names())

	// db.Exec(q).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Printf("%T %T", meta.ListValue...)

	// Output:
	// SELECT id,name FROM t;
	// *int *string
}

func ExampleMeta_BindWhere() {
	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)
		vals = []any{
			int(1000),
			string(`JohnDoe`),
		}
		idx int
	)

	idx = meta.BindWhere(``, vals[0])
	fmt.Printf("WHERE id=$%d\n", idx)

	idx = meta.BindWhere(``, vals[1])
	fmt.Printf("AND name=$%d\n", idx)

	fmt.Println(meta.ListWhereValue)

	// Output:
	// WHERE id=$1
	// AND name=$2
	// [1000 JohnDoe]
}

func ExampleMeta_Holders_mysql() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNameMysql, sql.DMLKindInsert)
		t    = Table{Name: `newname`, ID: 2}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	fmt.Printf("INSERT INTO t VALUES (%s);\n", meta.Holders())
	// Output:
	// INSERT INTO t VALUES (?,?);
}

func ExampleMeta_Holders_postgres() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindInsert)
		t    = Table{Name: `newname`, ID: 2}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	fmt.Printf("INSERT INTO t VALUES (%s);\n", meta.Holders())
	// Output:
	// INSERT INTO t VALUES ($1,$2);
}

func ExampleMeta_Names() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)
		t    = Table{}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	fmt.Printf("SELECT %s FROM t;\n", meta.Names())
	// Output:
	// SELECT id,name FROM t;
}

func ExampleMeta_Sub() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)
		t    = Table{}
		qid  = 1
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)
	meta.BindWhere(`id`, qid)

	var (
		metain = meta.Sub()
		qnames = []string{`hello`, `world`}
	)

	metain.BindWhere(``, qnames[0])
	metain.BindWhere(``, qnames[1])

	var q = fmt.Sprintf(`SELECT %s FROM t WHERE id=$1 OR name IN (%s);`,
		meta.Names(), metain.Holders())

	var qparams = sql.JoinValues(meta.ListWhereValue, metain.ListWhereValue)

	// db.QueryRow(q, qparams...).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Println(`SELECT #n=`, len(meta.ListValue))
	fmt.Println(`WHERE=`, meta.ListWhereValue)
	fmt.Println(`WHERE IN=`, metain.ListWhereValue)
	fmt.Println(`qparams=`, qparams)

	// Output:
	// SELECT id,name FROM t WHERE id=$1 OR name IN ($2,$3);
	// SELECT #n= 2
	// WHERE= [1]
	// WHERE IN= [hello world]
	// qparams= [1 hello world]
}

func ExampleMeta_UpdateFields() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindUpdate)
		t    = Table{
			ID:   2,
			Name: `world`,
		}
		qid   = 1
		qname = `hello`
	)

	meta.Bind(`id`, t.ID)
	meta.Bind(`name`, t.Name)
	meta.BindWhere(`id`, qid)
	meta.BindWhere(`AND name`, qname)

	var q = fmt.Sprintf(`UPDATE t SET %s WHERE %s;`, meta.UpdateFields(), meta.WhereFields())

	// db.Exec(q, meta.UpdateValues()...);

	fmt.Println(q)
	fmt.Println(`SET=`, meta.ListValue)
	fmt.Println(`WHERE=`, meta.ListWhereValue)
	fmt.Println(`Exec=`, meta.UpdateValues())

	// Output:
	// UPDATE t SET id=$1,name=$2 WHERE id=$3 AND name=$4;
	// SET= [2 world]
	// WHERE= [1 hello]
	// Exec= [2 world 1 hello]
}

func ExampleMeta_UpdateValues() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindUpdate)
		t    = Table{
			ID:   2,
			Name: `world`,
		}
		qid   = 1
		qname = `hello`
	)

	meta.Bind(`id`, t.ID)
	meta.Bind(`name`, t.Name)
	meta.BindWhere(`id`, qid)
	meta.BindWhere(`name`, qname)

	var q = fmt.Sprintf(`UPDATE t SET id=$%d,name=$%d WHERE id=$%d AND name=$%d;`, meta.Index...)

	// db.Exec(q, meta.UpdateValues()...);

	fmt.Println(q)
	fmt.Println(`Index=`, meta.Index)
	fmt.Println(`SET=`, meta.ListValue)
	fmt.Println(`WHERE=`, meta.ListWhereValue)
	fmt.Println(`Exec=`, meta.UpdateValues())

	// Output:
	// UPDATE t SET id=$1,name=$2 WHERE id=$3 AND name=$4;
	// Index= [1 2 3 4]
	// SET= [2 world]
	// WHERE= [1 hello]
	// Exec= [2 world 1 hello]
}

func ExampleMeta_WhereFields() {
	var meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)

	meta.BindWhere(`id`, 1000)
	meta.BindWhere(`AND name`, `share`)

	fmt.Printf(`SELECT * FROM t WHERE %s;`, meta.WhereFields())
	// Output:
	// SELECT * FROM t WHERE id=$1 AND name=$2;
}

func ExampleMeta_WhereHolders() {
	var meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)

	meta.BindWhere(`id`, 1000)
	meta.BindWhere(`name`, `share`)

	fmt.Printf(`SELECT * FROM t WHERE id IN (%s);`, meta.WhereHolders())
	// Output:
	// SELECT * FROM t WHERE id IN ($1,$2);
}

func ExampleMeta_deleteOnPostgresql() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta  = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindUpdate)
		qid   = 1
		qname = `hello`
	)

	meta.BindWhere(`id`, qid)
	meta.BindWhere(`OR name`, qname)

	var q = fmt.Sprintf(`DELETE FROM t WHERE %s;`, meta.WhereFields())

	// db.Exec(q, meta.ListWhereValue...)

	fmt.Println(q)
	fmt.Println(meta.ListWhereValue)

	// Output:
	// DELETE FROM t WHERE id=$1 OR name=$2;
	// [1 hello]
}

func ExampleMeta_insertOnPostgresql() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindInsert)
		t    = Table{
			ID:   1,
			Name: `hello`,
		}
	)

	meta.Bind(`id`, t.ID)
	meta.Bind(`name`, t.Name)

	var q = fmt.Sprintf(`INSERT INTO t (%s) VALUES (%s);`, meta.Names(), meta.Holders())

	// db.Exec(q, meta.ListValue...)

	fmt.Println(q)
	fmt.Println(meta.ListValue)

	// Output:
	// INSERT INTO t (id,name) VALUES ($1,$2);
	// [1 hello]
}

func ExampleMeta_selectOnPostgresql() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta  = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)
		t     = Table{}
		qid   = 1
		qname = `hello`
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)
	meta.BindWhere(`id`, qid)
	meta.BindWhere(`name`, qname)

	var q = fmt.Sprintf(`SELECT %s FROM t WHERE id=$1 OR name=$2;`, meta.Names())

	// db.QueryRow(q, meta.ListWhereValue...).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Println(len(meta.ListValue))
	fmt.Println(meta.ListWhereValue)

	// Output:
	// SELECT id,name FROM t WHERE id=$1 OR name=$2;
	// 2
	// [1 hello]
}

// Sometime the query need to be stiched piece by piece.
func ExampleMeta_subquery() {
	type Table struct {
		Name  string
		ID    int
		SubID int
	}

	var (
		meta  = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)
		id    = 1
		subid = 500
		t     Table
		qb    strings.Builder
		idx   int
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`sub_id`, &t.SubID)
	meta.Bind(`name`, &t.Name)

	fmt.Fprintf(&qb, `SELECT %s FROM t WHERE 1=1`, meta.Names())

	if id != 0 {
		idx = meta.BindWhere(``, id)
		fmt.Fprintf(&qb, ` AND id = $%d`, idx)
	}
	if subid != 0 {
		idx = meta.BindWhere(``, subid)
		fmt.Fprintf(&qb, ` AND sub_id = (SELECT id FROM u WHERE u.id = $%d);`, idx)
	}

	// db.Exec(qb.String(),meta.ListWhereValue...).Scan(meta.ListValue...)

	fmt.Println(qb.String())
	fmt.Println(meta.ListWhereValue)

	// Output:
	// SELECT id,sub_id,name FROM t WHERE 1=1 AND id = $1 AND sub_id = (SELECT id FROM u WHERE u.id = $2);
	// [1 500]
}

func ExampleMeta_subqueryWithIndex() {
	type Table struct {
		Name  string
		ID    int
		SubID int
	}

	var (
		meta  = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)
		id    = 1
		subid = 500
		t     Table
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`sub_id`, &t.SubID)
	meta.Bind(`name`, &t.Name)

	var qb strings.Builder

	fmt.Fprintf(&qb, `SELECT %s FROM t WHERE 1=1`, meta.Names())
	if id != 0 {
		qb.WriteString(` AND id = $%d`)
		meta.BindWhere(`id`, id)
	}
	if subid != 0 {
		qb.WriteString(` AND sub_id = (SELECT id FROM u WHERE u.id = $%d);`)
		meta.BindWhere(`sub_id`, subid)
	}

	var q = fmt.Sprintf(qb.String(), meta.Index...)

	// db.Exec(q, meta.ListWhereValue...).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Println(meta.Index)
	fmt.Println(meta.ListWhereValue)

	// Output:
	// SELECT id,sub_id,name FROM t WHERE 1=1 AND id = $1 AND sub_id = (SELECT id FROM u WHERE u.id = $2);
	// [1 2]
	// [1 500]
}
