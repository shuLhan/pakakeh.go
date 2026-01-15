<!-- SPDX-License-Identifier: BSD-3-Clause -->
<!-- SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info> -->

[![GoDoc](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula?status.svg)](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula)
[![Go Report Card](https://goreportcard.com/badge/git.sr.ht/~shulhan/pakakeh.go/lib/tabula)](https://goreportcard.com/report/git.sr.ht/~shulhan/pakakeh.go/lib/tabula)
![cover.run go](https://cover.run/go/git.sr.ht/~shulhan/pakakeh.go/lib/tabula.svg)

Package tabula is a Go library for working with rows, columns, or matrix
(table), or in another terms working with data set.

# Overview

Go's slice gave a flexible way to manage sequence of data in one type, but what
if you want to manage a sequence of value but with different type of data?
Or manage a bunch of values like a table?

You can use this library to manage sequence of value with different type
and manage data in two dimensional tuple.

## Terminology

Here are some terminologies that we used in developing this library, which may
help reader understand the internal and API.

Record is a single cell in row or column, or the smallest building block of
dataset.

Row is a horizontal representation of records in dataset.

Column is a vertical representation of records in dataset.
Each column has a unique name and has the same type data.

Dataset is a collection of rows and columns.

Given those definitions we can draw the representation of rows, columns, or
matrix:

            COL-0  COL-1 ...  COL-x
    ROW-0: record record ... record
    ROW-1: record record ... record
    ...
    ROW-y: record record ... record

## What make this package different from other dataset packages?

### Record Type

There are only three valid type in record: int64, float64, and string.

Each record is a pointer to interface value. Which means,

- Switching between rows to columns mode, or vice versa, is only a matter of
  pointer switching, no memory relocations.
- When using matrix mode, additional memory is used only to allocate slice, the
  record in each rows and columns is shared.

### Dataset Mode

Tabula has three mode for dataset: rows, columns, or matrix.

For example, given a table of data,

    col1,col2,col3
    a,b,c
    1,2,3

- When in "rows" mode, each line is saved in its own slice, resulting in Rows:

  ```
  Rows[0]: [a b c]
  Rows[1]: [1 2 3]
  ```

  Columns is used only to save record metadata: column name, type, flag and
  value space.

- When in "columns" mode, each line saved in columns, resulting in Columns:

  ```
  Columns[0]: {col1 0 0 [] [a 1]}
  Columns[1]: {col2 0 0 [] [b 2]}
  Columns[1]: {col3 0 0 [] [c 3]}
  ```

  Each column will contain metadata including column name, type, flag, and
  value space (all possible value that _may_ contain in column value).

  Rows in "columns" mode is empty.

- When in "matrix" mode, each record is saved both in row and column using
  shared pointer to record.

  Matrix mode consume more memory by allocating two slice in rows and columns,
  but give flexible way to manage records.

## Features

- **Switching between rows and columns mode**.

- [**Random pick rows with or without replacement**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#RandomPickRows).

- [**Random pick columns with or without replacement**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#RandomPickColumns).

- [**Select column from dataset by index**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#SelectColumnsByIdx).

- [**Sort columns by index**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#SortColumnsByIndex),
  or indirect sort.

- [**Split rows value by numeric**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#SplitRowsByNumeric).
  For example, given two numeric rows,

  ```
  A: {1,2,3,4}
  B: {5,6,7,8}
  ```

  if we split row by value 7, the data will splitted into left set

  ```
  A': {1,2}
  B': {5,6}
  ```

  and the right set would be

  ```
  A'': {3,4}
  B'': {7,8}
  ```

- [**Split rows by string**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#SplitRowsByCategorical).
  For example, given two rows,

  ```
  X: [A,B,A,B,C,D,C,D]
  Y: [1,2,3,4,5,6,7,8]
  ```

  if we split the rows with value set `[A,C]`, the data will splitted into left
  set which contain all rows that have A or C,

  ```
  		X': [A,A,C,C]
  		Y': [1,3,5,7]
  ```

  and the right set, excluded set, will contain all rows which is not A or C,

  ```
  		X'': [B,B,D,D]
  		Y'': [2,4,6,8]
  ```

- [**Select row where**](https://godoc.org/git.sr.ht/~shulhan/pakakeh.go/lib/tabula#SelectRowsWhere).
  Select row at column index x where their value is equal to y (an analogy to
  _select where_ in SQL).
  For example, given a rows of dataset,
  ```
  ROW-1: {1,A}
  ROW-2: {2,B}
  ROW-3: {3,A}
  ROW-4: {4,C}
  ```
  we can select row where the second column contain 'A', which result in,
  ```
  ROW-1: {1,A}
  ROW-3: {3,A}
  ```
