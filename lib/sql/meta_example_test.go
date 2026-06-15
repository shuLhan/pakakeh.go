// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package sql_test

import (
	"fmt"
	"slices"

	"kilabit.info/pakakeh.go/lib/sql"
)

func ExampleMeta_Bind() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNameMysql)
		t    = Table{}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	fmt.Println(meta.Select(`t`))
	fmt.Printf("%T %T", meta.ListValue...)

	// Output:
	// SELECT id,name FROM t
	// *int *string
}

func ExampleMeta_BindWhere() {
	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
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

func ExampleMeta_BulkInsert_postgres() {
	type Tag struct {
		Name string
		ID   int
	}
	type Document struct {
		Tags []Tag
		ID   int
	}

	meta := sql.NewMeta(sql.DriverNamePostgres)
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

	meta.BulkBindName(`document_id`, false)
	meta.BulkBindName(`tag_id`, true)

	for _, tag := range doc.Tags {
		meta.BulkBindValues(doc.ID, tag.ID)
	}

	q := meta.BulkInsert(`document_tag`)

	// db.Exec(q, meta.InsertValues()...)

	fmt.Println(q)
	fmt.Println(meta.InsertValues())

	// Output:
	// INSERT INTO document_tag (document_id,tag_id) VALUES ($1,$2),($1,$3);
	// [1 100 101]
}

func ExampleMeta_BulkInsert_nonPostgres() {
	type Tag struct {
		Name string
		ID   int
	}
	type Document struct {
		Tags []Tag
		ID   int
	}

	meta := sql.NewMeta(sql.DriverNameMysql)
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

	meta.BulkBindName(`document_id`, false)
	meta.BulkBindName(`tag_id`, true)

	for _, tag := range doc.Tags {
		meta.BulkBindValues(doc.ID, tag.ID)
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

func ExampleMeta_Delete() {
	qmeta := sql.NewMeta(sql.DriverNamePostgres)
	qmeta.BindWhere(``, `account_id`, `=`, 1000)
	qmeta.BindWhere(`AND`, `id`, `=`, 1)
	q := qmeta.Delete(`my_table`)
	fmt.Println(q)
	fmt.Println(qmeta.WhereValues())
	// Output:
	// DELETE FROM my_table WHERE  account_id = $1 AND id = $2
	// [1000 1]
}

func ExampleMeta_Holders_mysql() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNameMysql)
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
		meta = sql.NewMeta(sql.DriverNamePostgres)
		t    = Table{Name: `newname`, ID: 2}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	fmt.Printf("INSERT INTO t VALUES (%s);\n", meta.Holders())
	// Output:
	// INSERT INTO t VALUES ($1,$2);
}

func ExampleMeta_Insert() {
	qmeta := sql.NewMeta(sql.DriverNamePostgres)
	qmeta.Bind(`name`, `roana`)
	qmeta.Bind(`age`, 43)
	q := qmeta.Insert(`my_table`)
	fmt.Println(q)
	fmt.Println(qmeta.InsertValues())
	// Output:
	// INSERT INTO my_table (name,age) VALUES ($1,$2);
	// [roana 43]
}

func ExampleMeta_Names() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
		t    = Table{}
	)

	meta.Bind(`id`, &t.ID)
	meta.Bind(`name`, &t.Name)

	fmt.Printf("SELECT %s FROM t;\n", meta.Names())
	// Output:
	// SELECT id,name FROM t;
}

func ExampleMeta_Select() {
	var id int64
	var name string
	qmeta := sql.NewMeta(sql.DriverNamePostgres)
	qmeta.BindSelect(`id`, &id)
	qmeta.BindSelect(`name`, &name)
	q := qmeta.Select(`my_table`)
	fmt.Println(q)
	fmt.Printf("%T %T\n", qmeta.ScanHolders()...)
	// Output:
	// SELECT id,name FROM my_table
	// *int64 *string
}

func ExampleMeta_Select_withWhere() {
	var id int64
	var name string
	qmeta := sql.NewMeta(sql.DriverNamePostgres)
	qmeta.BindSelect(`id`, &id)
	qmeta.BindSelect(`name`, &name)
	qmeta.BindWhere(``, `id`, `>`, 10)
	q := qmeta.Select(`my_table`)
	fmt.Println(q)
	fmt.Println(qmeta.WhereValues())
	fmt.Printf("%T %T\n", qmeta.ScanHolders()...)
	// Output:
	// SELECT id,name FROM my_table WHERE  id > $1
	// [10]
	// *int64 *string
}

func ExampleMeta_Sub() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
		t    = Table{}
		qid  = 1
	)

	meta.BindSelect(`id`, &t.ID)
	meta.BindSelect(`name`, &t.Name)
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
	fmt.Println(`WHERE=`, meta.WhereValues())
	fmt.Println(`WHERE IN=`, metain.WhereValues())
	fmt.Println(`qparams=`, qparams)

	// Output:
	// SELECT id,name FROM t WHERE id=$1 OR name IN ($2,$3);
	// WHERE= [1]
	// WHERE IN= [hello world]
	// qparams= [1 hello world]
}

func ExampleMeta_Update() {
	qmeta := sql.NewMeta(sql.DriverNamePostgres)
	qmeta.Bind(`name`, `roana`)
	qmeta.Bind(`age`, 43)
	q := qmeta.Update(`my_table`)
	fmt.Println(q)
	fmt.Println(qmeta.UpdateValues())
	// Output:
	// UPDATE my_table SET name=$1,age=$2
	// [roana 43]
}

func ExampleMeta_Update_withWhere() {
	qmeta := sql.NewMeta(sql.DriverNamePostgres)
	qmeta.Bind(`name`, `roana`)
	qmeta.Bind(`age`, 44)
	qmeta.BindWhere(``, `id`, `=`, `1`)
	q := qmeta.Update(`my_table`)
	fmt.Println(q)
	fmt.Println(qmeta.UpdateValues())
	// Output:
	// UPDATE my_table SET name=$1,age=$2 WHERE  id = $3
	// [roana 44 1]
}

func ExampleMeta_UpdateFields() {
	type Table struct {
		Name string
		ID   int
	}

	var (
		meta = sql.NewMeta(sql.DriverNamePostgres)
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
		meta = sql.NewMeta(sql.DriverNamePostgres)
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
	var meta = sql.NewMeta(sql.DriverNamePostgres)

	meta.BindWhere(``, `id`, `=`, 1000)
	meta.BindWhere(`AND`, `name`, `=`, `share`)

	fmt.Printf("SELECT * FROM t WHERE %s;\n", meta.WhereFields())
	fmt.Println(meta.WhereValues())

	// Output:
	// SELECT * FROM t WHERE  id = $1 AND name = $2;
	// [1000 share]
}

func ExampleMeta_WhereHolders() {
	var meta = sql.NewMeta(sql.DriverNamePostgres)

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
		meta  = sql.NewMeta(sql.DriverNamePostgres)
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
		meta = sql.NewMeta(sql.DriverNamePostgres)
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
		meta  = sql.NewMeta(sql.DriverNamePostgres)
		t     = Table{}
		qid   = 1
		qname = `hello`
	)

	meta.BindSelect(`id`, &t.ID)
	meta.BindSelect(`name`, &t.Name)
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

// Example of creating parent and sub-query.
func ExampleMeta_subquery() {
	type Table struct {
		Name  string
		ID    int
		SubID int
	}

	var (
		meta  = sql.NewMeta(sql.DriverNamePostgres)
		id    = 1
		subid = 500
		t     Table
	)

	meta.BindSelect(`id`, &t.ID)
	meta.BindSelect(`sub_id`, &t.SubID)
	meta.BindSelect(`name`, &t.Name)
	meta.BindWhere(``, `id`, `=`, id)

	subq := meta.Sub()
	subq.BindSelect(`id`, nil)
	subq.BindWhere(``, `u.id`, `=`, subid)

	q := fmt.Sprintf(`%s AND sub_id = (%s)`, meta.Select(`t`), subq.Select(`u`))
	vals := slices.Concat(meta.WhereValues(), subq.WhereValues())
	fmt.Println(q)
	fmt.Println(vals)

	// Output:
	// SELECT id,sub_id,name FROM t WHERE  id = $1 AND sub_id = (SELECT id FROM u WHERE  u.id = $2)
	// [1 500]
}
