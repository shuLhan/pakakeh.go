// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package sql_test

import (
	"fmt"
	"slices"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/sql"
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
	)

	opWhere := meta.BindWhere(``, `id`, `=`, vals[0])
	fmt.Printf("%s\n", opWhere.String())

	opWhere = meta.BindWhere(`AND`, `name`, `=`, vals[1])
	fmt.Printf("%s\n", opWhere.String())

	fmt.Println(meta.WhereValues())

	// Output:
	// id = $1
	//  AND name = $2
	// [1000 JohnDoe]
}

func ExampleMeta_bulkInsert_postgres() {
	type Tag struct {
		Name string
		ID   int
	}
	type Document struct {
		Tags []Tag
		ID   int
	}

	meta := sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindInsert)
	doc := Document{
		ID: 1,
		Tags: []Tag{{
			ID:   100,
			Name: `tag A`,
		}, {
			ID:   101,
			Name: `tag B`,
		}},
	}

	meta.BindName(`document_id`, false)
	meta.BindName(`tag_id`, true)

	for _, tag := range doc.Tags {
		meta.BindValues(doc.ID, tag.ID)
	}

	var q = fmt.Sprintf(`INSERT INTO document_tag (%s) VALUES %s;`,
		meta.Names(), meta.BulkHolders())

	// db.Exec(q, meta.InsertValues()...)

	fmt.Println(q)
	fmt.Println(meta.InsertValues())

	// Output:
	// INSERT INTO document_tag (document_id,tag_id) VALUES ($1,$2),($1,$3);
	// [1 100 101]
}

func ExampleMeta_bulkInsert_nonPostgres() {
	type Tag struct {
		Name string
		ID   int
	}
	type Document struct {
		Tags []Tag
		ID   int
	}

	meta := sql.NewMeta(sql.DriverNameMysql, sql.DMLKindInsert)
	doc := Document{
		ID: 1,
		Tags: []Tag{{
			ID:   100,
			Name: `tag A`,
		}, {
			ID:   101,
			Name: `tag B`,
		}},
	}

	meta.BindName(`document_id`, false)
	meta.BindName(`tag_id`, true)

	for _, tag := range doc.Tags {
		meta.BindValues(doc.ID, tag.ID)
	}

	var q = fmt.Sprintf(`INSERT INTO document_tag (%s) VALUES %s;`,
		meta.Names(), meta.BulkHolders())

	// db.Exec(q, meta.InsertValues()...)

	fmt.Println(q)
	fmt.Println(meta.InsertValues())

	// Output:
	// INSERT INTO document_tag (document_id,tag_id) VALUES (?,?),(?,?);
	// [1 100 1 101]
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
	meta.BindWhere(``, `id`, `=`, qid)

	var (
		metain = meta.Sub()
		qnames = []string{`hello`, `world`}
	)

	metain.BindWhere(``, ``, ``, qnames[0])
	metain.BindWhere(``, ``, ``, qnames[1])

	var q = fmt.Sprintf(`SELECT %s FROM t WHERE id=$1 OR name IN (%s);`,
		meta.Names(), metain.WhereHolders())

	var qparams = slices.Concat(meta.WhereValues(), metain.WhereValues())

	// db.QueryRow(q, qparams...).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Println(`SELECT #n=`, len(meta.ListValue))
	fmt.Println(`WHERE=`, meta.WhereValues())
	fmt.Println(`WHERE IN=`, metain.WhereValues())
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
	meta.BindWhere(``, `id`, `=`, qid)
	meta.BindWhere(`AND`, `name`, `=`, qname)

	var q = fmt.Sprintf(`UPDATE t SET %s WHERE %s;`,
		meta.UpdateFields(), meta.WhereFields())

	// db.Exec(q, meta.UpdateValues()...);

	fmt.Println(q)
	fmt.Println(`UpdateValues=`, meta.UpdateValues())

	// Output:
	// UPDATE t SET id=$1,name=$2 WHERE  id = $3 AND name = $4;
	// UpdateValues= [2 world 1 hello]
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
	meta.BindWhere(``, `id`, `=`, qid)
	meta.BindWhere(`AND`, `name`, `=`, qname)

	var q = fmt.Sprintf(`UPDATE t SET %s WHERE %s;`,
		meta.UpdateFields(), meta.WhereFields())

	// db.Exec(q, meta.UpdateValues()...);

	fmt.Println(q)
	fmt.Println(`UpdateValues=`, meta.UpdateValues())

	// Output:
	// UPDATE t SET id=$1,name=$2 WHERE  id = $3 AND name = $4;
	// UpdateValues= [2 world 1 hello]
}

func ExampleMeta_WhereFields() {
	var meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)

	meta.BindWhere(``, `id`, `=`, 1000)
	meta.BindWhere(`AND`, `name`, `=`, `share`)

	fmt.Printf("SELECT * FROM t WHERE %s;\n", meta.WhereFields())
	fmt.Println(meta.WhereValues())

	// Output:
	// SELECT * FROM t WHERE  id = $1 AND name = $2;
	// [1000 share]
}

func ExampleMeta_WhereHolders() {
	var meta = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindSelect)

	meta.BindWhere(``, ``, ``, 1000)
	meta.BindWhere(``, ``, ``, `share`)

	fmt.Printf("SELECT * FROM t WHERE id IN (%s);\n", meta.WhereHolders())
	fmt.Println(meta.WhereValues())

	// Output:
	// SELECT * FROM t WHERE id IN ($1,$2);
	// [1000 share]
}

func ExampleMeta_deleteOnPostgresql() {
	var (
		meta  = sql.NewMeta(sql.DriverNamePostgres, sql.DMLKindUpdate)
		qid   = 1
		qname = `hello`
	)

	meta.BindWhere(``, `id`, `=`, qid)
	meta.BindWhere(`OR`, `name`, `=`, qname)

	var q = fmt.Sprintf(`DELETE FROM t WHERE %s;`, meta.WhereFields())

	// db.Exec(q, meta.WhereValues()...)

	fmt.Println(q)
	fmt.Println(meta.WhereValues())

	// Output:
	// DELETE FROM t WHERE  id = $1 OR name = $2;
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
	meta.BindWhere(``, `id`, `=`, qid)
	meta.BindWhere(`OR`, `name`, `=`, qname)

	var q = fmt.Sprintf(`SELECT %s FROM t WHERE %s;`, meta.Names(), meta.WhereFields())

	// db.QueryRow(q, meta.WhereValues()...).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Println(`WHERE=`, meta.WhereValues())
	fmt.Println(len(meta.ListValue))

	// Output:
	// SELECT id,name FROM t WHERE  id = $1 OR name = $2;
	// WHERE= [1 hello]
	// 2
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
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`sub_id`, &t.SubID)
	meta.Bind(`name`, &t.Name)

	fmt.Fprintf(&qb, `SELECT %s FROM t WHERE 1=1`, meta.Names())

	if id != 0 {
		op := meta.BindWhere(`AND`, `id`, `=`, id)
		qb.WriteString(op.String())
	}
	if subid != 0 {
		op := meta.BindWhere(``, `u.id`, `=`, subid)
		fmt.Fprintf(&qb, ` AND sub_id = (SELECT id FROM u WHERE %s);`,
			op.String())
	}

	// db.Exec(qb.String(),meta.WhereValues()...).Scan(meta.ListValue...)

	fmt.Println(qb.String())
	fmt.Println(meta.WhereValues())

	// Output:
	// SELECT id,sub_id,name FROM t WHERE 1=1 AND id = $1 AND sub_id = (SELECT id FROM u WHERE  u.id = $2);
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
		meta.BindWhere(``, `id`, `=`, id)
	}
	if subid != 0 {
		qb.WriteString(` AND sub_id = (SELECT id FROM u WHERE u.id = $%d);`)
		meta.BindWhere(``, `u.id`, `=`, subid)
	}

	var q = fmt.Sprintf(qb.String(), meta.Index...)

	// db.Exec(q, meta.WhereValues()...).Scan(meta.ListValue...)

	fmt.Println(q)
	fmt.Println(meta.Index)
	fmt.Println(meta.WhereValues())

	// Output:
	// SELECT id,sub_id,name FROM t WHERE 1=1 AND id = $1 AND sub_id = (SELECT id FROM u WHERE u.id = $2);
	// [1 2]
	// [1 500]
}
