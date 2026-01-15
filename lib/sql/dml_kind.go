// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package sql

// DMLKind define the kind for Data Manipulation Language (DML).
type DMLKind string

// List of valid DMLKind.
const (
	DMLKindDelete DMLKind = `DELETE`
	DMLKindInsert DMLKind = `INSERT`
	DMLKindSelect DMLKind = `SELECT`
	DMLKindUpdate DMLKind = `UPDATE`
)
