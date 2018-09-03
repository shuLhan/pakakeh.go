// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file

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
// Open the file in path for reading.
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
// ReadUntil read the content of file until one of separator found, or until
// it reach the terminator character, or until EOF.
// The content will be returned along the status of termination.
// If terminator or EOF found, the returned isTerm value will be true,
// otherwise it will be false.
//
func (r *Reader) ReadUntil(seps []byte, terms []byte) (b []byte, isTerm bool, c byte) {
	for r.p < len(r.v) {
		for x := 0; x < len(terms); x++ {
			if r.v[r.p] == terms[x] {
				c = r.v[r.p]
				r.p++
				isTerm = true
				return
			}
		}
		for x := 0; x < len(seps); x++ {
			if r.v[r.p] == seps[x] {
				c = r.v[r.p]
				r.p++
				return
			}
		}
		b = append(b, r.v[r.p])
		r.p++
	}
	return
}

//
// SkipSpace read until no white spaces found and return the first byte that
// is not white spaces.
// On EOF, it will return 0.
//
func (r *Reader) SkipSpace() byte {
	for r.p < len(r.v) {
		if libbytes.IsSpace(r.v[r.p]) {
			r.p++
			continue
		}
		break
	}
	if r.p == len(r.v) {
		return 0
	}
	return r.v[r.p]
}

//
// SkipHorizontalSpace read until no space, carriage return, or tab occured on
// buffer.
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
func (r *Reader) SkipUntil(seps []byte) {
	for r.p < len(r.v) {
		for x := 0; x < len(seps); x++ {
			if r.v[r.p] == seps[x] {
				r.p++
				return
			}
		}
		r.p++
	}
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
