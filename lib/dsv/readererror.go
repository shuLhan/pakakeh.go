// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"fmt"
)

const (
	_ = iota
	// EReadMissLeftQuote read error when no left-quote found on line.
	EReadMissLeftQuote
	// EReadMissRightQuote read error when no right-quote found on line.
	EReadMissRightQuote
	// EReadMissSeparator read error when no separator found on line.
	EReadMissSeparator
	// EReadLine error when reading line from file.
	EReadLine
	// EReadEOF error which indicated end-of-file.
	EReadEOF
	// ETypeConversion error when converting type from string to numeric or
	// vice versa.
	ETypeConversion
)

//
// ReaderError to handle error data and message.
//
type ReaderError struct {
	// T define type of error.
	T int
	// Func where error happened
	Func string
	// What cause the error?
	What string
	// Line define the line which cause error
	Line string
	// Pos character position which cause error
	Pos int
	// N line number
	N int
}

//
// Error to string.
//
func (e *ReaderError) Error() string {
	return fmt.Sprintf("dsv.Reader.%-20s [%d:%d]: %-30s data:|%s|", e.Func, e.N,
		e.Pos, e.What, e.Line)
}
