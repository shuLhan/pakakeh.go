// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>
// in the LICENSE file.

package tabula

var (
	testColTypes = []int{
		TInteger,
		TInteger,
		TInteger,
		TString,
	}

	testColNames = []string{"int01", "int02", "int03", "class"}

	// Testing data and function for Rows and MapRows
	rowsData = [][]string{
		{"1", "5", "9", "+"},
		{"2", "6", "0", "-"},
		{"3", "7", "1", "-"},
		{"4", "8", "2", "+"},
	}

	testClassIdx = 3

	rowsExpect = []string{
		"&[1 5 9 +]",
		"&[2 6 0 -]",
		"&[3 7 1 -]",
		"&[4 8 2 +]",
	}

	groupByExpect = "[{+ &[1 5 9 +]&[4 8 2 +]} {- &[2 6 0 -]&[3 7 1 -]}]"
)

func initRows() (rows Rows, e error) {
	for i := range rowsData {
		l := len(rowsData[i])
		row := make(Row, 0)

		for j := 0; j < l; j++ {
			rec, e := NewRecordBy(rowsData[i][j],
				testColTypes[j])

			if nil != e {
				return nil, e
			}

			row = append(row, rec)
		}

		rows.PushBack(&row)
	}
	return rows, nil
}
