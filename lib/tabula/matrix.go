// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>

package tabula

// Matrix is a combination of columns and rows.
type Matrix struct {
	Columns *Columns
	Rows    *Rows
}
