// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package sql

// param contains column name and properties for select, insert, or update.
type param struct {
	name   string
	holder string
	unique bool
}
