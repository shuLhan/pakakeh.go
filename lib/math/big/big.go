// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package big extends the capabilities of standard "math/big" package by
// adding custom global precision to Float and Rat; and global rounding mode,
// and bits precision to Float.
package big

// DefaultDigitPrecision define the default number of digits after decimal
// point which affect the return of String() and MarshalJSON() methods.
//
// A zero value of digit precision mean is it will use the default output of
// 'f' format.
//
// One should change this value before using the new extended Float or Rat in
// the program.
var DefaultDigitPrecision = 8

// MarshalJSONAsString define the default return behaviour of MarshalJSON().
// If its true (the default) the returned JSON format will encapsulated in
// double quote, as string instead of as numeric.
var MarshalJSONAsString = true
