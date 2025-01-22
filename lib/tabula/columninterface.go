// SPDX-FileCopyrightText: 2017 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package tabula

// ColumnInterface define an interface for working with Column.
type ColumnInterface interface {
	SetType(tipe int)
	SetName(name string)

	GetType() int
	GetName() string

	SetRecords(recs *Records)

	Interface() any
}
