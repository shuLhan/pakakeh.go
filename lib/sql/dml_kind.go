// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

// DmlKind define the kind for Data Manipulation Language (DML).
type DmlKind string

// List of valid DmlKind.
const (
	DmlKindDelete DmlKind = `DELETE`
	DmlKindInsert DmlKind = `INSERT`
	DmlKindSelect DmlKind = `SELECT`
	DmlKindUpdate DmlKind = `UPDATE`
)
