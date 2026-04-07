// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package sql

// OpWhere stores the operation for WHERE conditions from calling
// [Meta.BindWhere].
type OpWhere struct {
	val    any
	logic  string
	column string
	comp   string
	holder string
}

func (op OpWhere) String() string {
	// The space between column is required for "IS" like comparison.
	if op.logic == `` {
		return ` ` + op.column + ` ` + op.comp + ` ` + op.holder
	}
	return ` ` + op.logic + ` ` + op.column + ` ` + op.comp + ` ` + op.holder
}
