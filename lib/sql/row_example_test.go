package sql

import (
	"fmt"
	"strings"
)

func ExampleRow_ExtractSQLFields() {
	row := Row{
		"col_3": "'update'",
		"col_2": 1,
		"col_1": true,
	}
	names, holders, values := row.ExtractSQLFields("?")
	fnames := strings.Join(names, ",")
	fholders := strings.Join(holders, ",")
	q := `INSERT INTO table (` + fnames + `) VALUES (` + fholders + `)`
	fmt.Printf("Query: %s\n", q)

	// err := db.Exec(q, values...)
	fmt.Println(values)

	names, holders, values = row.ExtractSQLFields("postgres")
	fnames = strings.Join(names, ",")
	fholders = strings.Join(holders, ",")
	q = `INSERT INTO table (` + fnames + `) VALUES (` + fholders + `)`
	fmt.Printf("Query for PostgreSQL: %s\n", q)

	// err := db.Exec(q, values...)
	fmt.Println(values)

	// Output:
	// Query: INSERT INTO table (col_1,col_2,col_3) VALUES (?,?,?)
	// [true 1 'update']
	// Query for PostgreSQL: INSERT INTO table (col_1,col_2,col_3) VALUES ($1,$2,$3)
	// [true 1 'update']
}
