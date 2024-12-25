// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

// ApoOp define the write operation of data.
type ApoOp byte

// List of possible Apo write operation.
const (
	ApoOpInsert  ApoOp = 0 // Default operation.
	ApoOpUpdate        = 1
	ApoOpReplace       = 2
	ApoOpDelete        = 4
)

// ApoMeta define the metadata for each data.
type ApoMeta struct {
	// At contains the timestamp with nanoseconds, in UTC timezone, when
	// Write called.
	At int64

	// Kind define the type of data.
	// The value of this field is defined by user, to know type of data
	// stored for reading later.
	Kind int32

	// Op define the write operation, including: insert, update,
	// replace, or delete.
	Op ApoOp
}
