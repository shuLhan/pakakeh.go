// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//
// Package tabula is a Go library for working with rows, columns, or matrix
// (table), or in another terms working with data set.
//
// Introduction
//
// Go's slice gave a flexible way to manage sequence of data in one type, but
// what if you want to manage a sequence of value but with different type of
// data?  Or manage a bunch of values like a table?
//
// You can use this library to manage sequence of value with different type
// and manage data in two dimensional tuple.
//
// Terminology
//
// Here are some terminologies that we used in developing this library, which
// may help reader understand the internal and API.
//
// Record is a single cell in row or column, or the smallest building block of
// dataset.
//
// Row is a horizontal representation of records in dataset.
//
// Column is a vertical representation of records in dataset.
// Each column has a unique name and has the same type data.
//
// Dataset is a collection of rows and columns.
//
// Given those definitions we can draw the representation of rows, columns, or
// matrix:
//
// 	        COL-0  COL-1 ...  COL-x
// 	ROW-0: record record ... record
// 	ROW-1: record record ... record
// 	...
// 	ROW-y: record record ... record
//
// Record Type
//
// There are only three valid type in record: int64, float64, and string.
//
// Dataset Mode
//
// Tabula has three mode for dataset: rows, columns, or matrix.
//
// For example, given a table of data,
//
//     col1,col2,col3
//     a,b,c
//     1,2,3
//
// "rows" mode is where each line saved in its own slice, resulting in Rows:
//
//     Rows[0]: [a b c]
//     Rows[1]: [1 2 3]
//
// "columns" mode is where each line saved by columns, resulting in Columns:
//
//     Columns[0]: {col1 0 0 [] [a 1]}
//     Columns[1]: {col2 0 0 [] [b 2]}
//     Columns[1]: {col3 0 0 [] [c 3]}
//
// Unlike rows mode, each column contain metadata including column name, type,
// flag, and value space (all possible value that _may_ contain in column
// value).
//
// "matrix" mode is where each record saved both in row and column.
//
// Matrix mode consume more memory but give a flexible way to manage records.
//
//
package tabula
