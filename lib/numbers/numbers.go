// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package numbers provide miscellaneous functions for working with integer,
// float, slice of integer, and slice of floats.
//
// # Features
//
// List of current features,
//
//   - sort slice of floats using in-place mergesort algorithm.
//   - sort slice of integer/floats by predefined index
//   - count number of value occurrence in slice of integer/float
//   - find minimum or maximum value in slice of integer/float
//   - sum slice of integer/float
package numbers

// SortThreshold when the data less than SortThreshold, insertion sort
// will be used to replace mergesort.
const SortThreshold = 7
