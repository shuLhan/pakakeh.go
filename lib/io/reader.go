// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io/ioutil"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// Reader represent a buffered reader that use an index to move through slice
// of bytes.
//
// The following illustration show the uses of each fields,
//
//	+-+-+-+-+-+
//	| | | | | | <= r.V
//	+-+-+-+-+-+
//	   ^
//	   |
//	  r.X
//
type Reader struct {
	X int    // X contains the current index of readed buffer.
	V []byte // V contains the buffer.
}

//
// NewReader open the file in path for reading.
//
func NewReader(path string) (*Reader, error) {
	var err error
	r := new(Reader)

	r.V, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//
// Init initialize reader buffer from slice of byte.
//
func (r *Reader) Init(src []byte) {
	r.X = 0
	r.V = src
}

//
// Current byte at index position or 0 if EOF.
//
func (r *Reader) Current() byte {
	if r.X == len(r.V) {
		return 0
	}
	return r.V[r.X]
}

//
// ReadLine read one line including the line feed '\n' character.
//
func (r *Reader) ReadLine() (line []byte) {
	if r.X == len(r.V) {
		return
	}

	start := r.X
	for r.X < len(r.V) {
		c := r.V[r.X]
		if c == '\n' {
			r.X++
			line = r.V[start:r.X]
			return
		}
		r.X++
	}
	line = r.V[start:]
	return line
}

//
// ReadUntil read the content of buffer until one of separator found,
// or until one of terminator character found, or until EOF.
// If terminator found, the returned isTerm value will be true, and c
// value will be the character that cause the termination.
//
func (r *Reader) ReadUntil(seps, terms []byte) (tok []byte, isTerm bool, c byte) {
	start := r.X
	for r.X < len(r.V) {
		c = r.V[r.X]
		for x := 0; x < len(terms); x++ {
			if c == terms[x] {
				tok = r.V[start:r.X]
				r.X++
				return tok, true, c
			}
		}
		for x := 0; x < len(seps); x++ {
			if c == seps[x] {
				tok = r.V[start:r.X]
				r.X++
				return tok, false, c
			}
		}
		r.X++
	}
	tok = r.V[start:]
	return tok, false, 0
}

//
// Rest return the rest of unreaded buffer.
//
func (r *Reader) Rest() []byte {
	return r.V[r.X:]
}

//
// ScanInt64 convert textual representation of number into int64 and return
// it.
// Any spaces before actual reading of text will be ignored.
// The number may prefixed with '-' or '+', if its '-', the returned value
// must be negative.
//
// On success, c is non digit character that terminate scan, if its 0, its
// mean EOF.
//
func (r *Reader) ScanInt64() (n int64, c byte) {
	var min int64 = 1
	if len(r.V) == r.X {
		return
	}

	for ; r.X < len(r.V); r.X++ {
		c = r.V[r.X]
		if !libbytes.IsSpace(c) {
			break
		}
	}
	if c == '-' {
		min = -1
		r.X++
	} else if c == '+' {
		r.X++
	}
	for r.X < len(r.V) {
		c = r.V[r.X]
		if !libbytes.IsDigit(c) {
			break
		}
		c -= '0'
		n *= 10
		n += int64(c)
		r.X++
	}
	n *= min
	if r.X == len(r.V) {
		return n, 0
	}

	return n, c
}

//
// SkipN skip reading n bytes from buffer and return true if EOF.
//
func (r *Reader) SkipN(n int) bool {
	r.X += n
	if r.X >= len(r.V) {
		r.X = len(r.V)
		return true
	}
	return false
}

//
// SkipSpace read until no white spaces found and return the first byte that
// is not white spaces.
// On EOF, it will return 0.
//
func (r *Reader) SkipSpace() (c byte) {
	for r.X < len(r.V) {
		c = r.V[r.X]
		if libbytes.IsSpace(c) {
			r.X++
			continue
		}
		return c
	}
	return 0
}

//
// SkipHorizontalSpace read until no space, carriage return, or tab occurred
// on buffer.
// On EOF it will return 0.
//
func (r *Reader) SkipHorizontalSpace() (n int, c byte) {
	for r.X < len(r.V) {
		if r.V[r.X] == '\t' || r.V[r.X] == '\r' || r.V[r.X] == ' ' {
			r.X++
			n++
			continue
		}
		break
	}
	if r.X == len(r.V) {
		return n, 0
	}
	return n, r.V[r.X]

}

//
// SkipUntil skip reading content until one of separator found or EOF.
//
func (r *Reader) SkipUntil(seps []byte) (c byte) {
	for r.X < len(r.V) {
		c = r.V[r.X]
		for x := 0; x < len(seps); x++ {
			if c == seps[x] {
				r.X++
				return c
			}
		}
		r.X++
	}
	return 0
}

//
// SkipLine skip reading content until newline.
//
func (r *Reader) SkipLine() {
	for r.X < len(r.V) {
		if r.V[r.X] == '\n' {
			r.X++
			return
		}
		r.X++
	}
}

//
// String return all unreaded content as string.
//
func (r *Reader) String() string {
	return string(r.V[r.X:])
}

//
// Unread the buffer N characters and return the character its pointed to.
// If N greater than buffer length, it will reset the pointer index back to
// zero.
//
func (r *Reader) UnreadN(n int) byte {
	if n > r.X {
		r.X = 0
	} else {
		r.X -= n
	}
	return r.V[r.X]
}
