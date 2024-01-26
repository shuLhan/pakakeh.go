// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
