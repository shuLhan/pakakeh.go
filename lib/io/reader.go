// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"io/ioutil"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

//
// Reader for file with delimited separated values.
//
type Reader struct {
	p int
	v []byte
}

//
// NewReader open the file in path for reading.
//
func NewReader(path string) (*Reader, error) {
	var err error
	r := new(Reader)

	r.v, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//
// Init initialize reader buffer from string.  This is an alternative of
// NewReader without opening and reading from file.
//
func (r *Reader) Init(src string) {
	r.p = 0
	r.v = []byte(src)
}

//
// InitBytes initialize reader buffer from slice of byte.
//
func (r *Reader) InitBytes(src []byte) {
	r.p = 0
	r.v = src
}

//
// Current return the byte at current index position or 0 if EOB
// (End-Of-Buffer).
//
func (r *Reader) Current() byte {
	if r.p == len(r.v) {
		return 0
	}
	return r.v[r.p]
}

//
// ReadUntil read the content of file until one of separator found, or until
// it reach the terminator character, or until EOF.
// The content will be returned along the status of termination.
// If terminator or EOF found, the returned isTerm value will be true,
// otherwise it will be false.
//
func (r *Reader) ReadUntil(seps, terms []byte) (b []byte, isTerm bool, c byte) {
	start := r.p
	for r.p < len(r.v) {
		for x := 0; x < len(terms); x++ {
			if r.v[r.p] == terms[x] {
				b = r.v[start:r.p]
				c = r.v[r.p]
				r.p++
				isTerm = true
				return
			}
		}
		for x := 0; x < len(seps); x++ {
			if r.v[r.p] == seps[x] {
				b = r.v[start:r.p]
				c = r.v[r.p]
				r.p++
				return
			}
		}
		r.p++
	}
	b = r.v[start:]
	return
}

//
// Rest return the rest of unreaded buffer.
//
func (r *Reader) Rest() []byte {
	return r.v[r.p:]
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
	if len(r.v) == r.p {
		return
	}

	for ; r.p < len(r.v); r.p++ {
		c = r.v[r.p]
		if !libbytes.IsSpace(c) {
			break
		}
	}
	if c == '-' {
		min = -1
		r.p++
	} else if c == '+' {
		r.p++
	}
	for r.p < len(r.v) {
		c = r.v[r.p]
		if !libbytes.IsDigit(c) {
			break
		}
		c = c - '0'
		n *= 10
		n += int64(c)
		r.p++
	}
	n *= min
	if r.p == len(r.v) {
		return n, 0
	}

	return n, c
}

//
// SkipN skip reading n bytes from buffer and return true if EOF.
//
func (r *Reader) SkipN(n int) bool {
	r.p += n
	if r.p >= len(r.v) {
		r.p = len(r.v)
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
	for r.p < len(r.v) {
		c = r.v[r.p]
		if libbytes.IsSpace(c) {
			r.p++
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
	for r.p < len(r.v) {
		if r.v[r.p] == '\t' || r.v[r.p] == '\r' || r.v[r.p] == ' ' {
			r.p++
			n++
			continue
		}
		break
	}
	if r.p == len(r.v) {
		return n, 0
	}
	return n, r.v[r.p]

}

//
// SkipUntil skip reading content until one of separator found or EOF.
//
func (r *Reader) SkipUntil(seps []byte) (c byte) {
	for r.p < len(r.v) {
		c = r.v[r.p]
		for x := 0; x < len(seps); x++ {
			if c == seps[x] {
				r.p++
				return c
			}
		}
		r.p++
	}
	return 0
}

//
// SkipUntilNewline skip reading content until newline.
//
func (r *Reader) SkipUntilNewline() {
	for r.p < len(r.v) {
		if r.v[r.p] == '\n' {
			r.p++
			return
		}
		r.p++
	}
}

//
// String return all unreaded content as string.
//
func (r *Reader) String() string {
	return string(r.v[r.p:])
}

//
// Unread the buffer N characters and return the character its pointed to.
// If N greater than, it will reset the pointer index back to zero.
//
func (r *Reader) UnreadN(n int) byte {
	if n > r.p {
		r.p = 0
	} else {
		r.p -= n
	}
	return r.v[r.p]
}
