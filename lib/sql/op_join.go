// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package sql

import "fmt"

type opJoin struct {
	mode      string
	onTable   string
	alias     string
	onColumn  string
	refColumn string
}

func (op opJoin) String() string {
	if op.alias == `` {
		return fmt.Sprintf(` %s %s ON %s = %s`, op.mode, op.onTable,
			op.onColumn, op.refColumn)
	}
	return fmt.Sprintf(` %s %s AS %s ON %s = %s`, op.mode, op.onTable,
		op.alias, op.onColumn, op.refColumn)
}
