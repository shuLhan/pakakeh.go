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
	names, holders, values := row.ExtractSQLFields(DefaultPlaceHolder)
	fnames := strings.Join(names, ",")
	fholders := strings.Join(holders, ",")

	q := `INSERT INTO table (` + fnames + `) VALUES (` + fholders + `)`
	fmt.Println(q)

	// err := db.Exec(q, values...)
	fmt.Println(values)

	//Output:
	//INSERT INTO table (col_1,col_2,col_3) VALUES (?,?,?)
	//[true 1 'update']
}
