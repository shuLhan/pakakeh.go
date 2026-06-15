// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package sql

// opOrderBy stores the operation for ORDER BY statement.
type opOrderBy struct {
	name  string
	order string
}

func (op opOrderBy) String() string {
	return op.name + ` ` + op.order
}
