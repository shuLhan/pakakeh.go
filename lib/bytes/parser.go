// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes

// Parser implement tokenize parser for stream of byte using one or more
// delimiters as separator between token.
type Parser struct {
	content []byte // The content to be parsed.
	delims  []byte // List of delimiters.
	x       int    // The position of current read.
	size    int    // The length of content.
}

// NewParser create new Parser to parse content using delims as initial
// delimiters.
func NewParser(content, delims []byte) (bp *Parser) {
	bp = &Parser{
		content: content,
		delims:  delims,
		size:    len(content),
	}
	return bp
}

// Read read a token until one of the delimiters found.
// If one of delimiter match, it will return it as d.
// When end of content encountered, the returned token may be not empty but
// the d will be zero.
func (bp *Parser) Read() (token []byte, d byte) {
	var c byte
	for bp.x < bp.size {
		c = bp.content[bp.x]
		for _, d = range bp.delims {
			if d == c {
				bp.x++
				return token, d
			}
		}
		token = append(token, c)
		bp.x++
	}
	return token, 0
}

// SkipLine skip all characters until new line.
// It will return 0 if EOF.
func (bp *Parser) SkipLine() (c byte) {
	for bp.x < bp.size {
		c = bp.content[bp.x]
		if c == '\n' {
			bp.x++
			return c
		}
		bp.x++
	}
	return 0
}
