// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bytes

import (
	"bytes"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

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

// AddDelimiters add another delimiters to the current parser.
func (bp *Parser) AddDelimiters(delims []byte) {
	bp.delims = append(bp.delims, delims...)
}

// Delimiters return the copy of current delimiters.
func (bp *Parser) Delimiters() []byte {
	return bytes.Clone(bp.delims)
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

// ReadLine read until it found new line ('\n') or end of content, ignoring
// all delimiters.
// The returned line will not contain '\n'.
func (bp *Parser) ReadLine() (line []byte, c byte) {
	for bp.x < bp.size {
		c = bp.content[bp.x]
		if c == '\n' {
			bp.x++
			return line, c
		}
		line = append(line, c)
		bp.x++
	}
	return line, 0
}

// ReadN read exactly n characters ignoring the delimiters.
// It will return the token and the character after n or 0 if end-of-content.
func (bp *Parser) ReadN(n int) (token []byte, d byte) {
	var (
		c     byte
		count int
	)
	for bp.x < bp.size {
		c = bp.content[bp.x]
		if count >= n {
			return token, c
		}
		token = append(token, c)
		count++
		bp.x++
	}
	return token, 0
}

// ReadNoSpace read the next token by ignoring the leading spaces, even if its
// one of the delimiter.
// The returned token will have no trailing spaces.
func (bp *Parser) ReadNoSpace() (token []byte, d byte) {
	var c byte

	// Ignore leading spaces.
	for ; bp.x < bp.size; bp.x++ {
		c = bp.content[bp.x]
		if !ascii.IsSpace(c) {
			break
		}
	}

	for ; bp.x < bp.size; bp.x++ {
		c = bp.content[bp.x]
		for _, d = range bp.delims {
			if d == c {
				bp.x++
				goto out
			}
		}
		token = append(token, c)
	}
	d = 0

out:
	// Remove trailing spaces.
	var x int
	for x = len(token) - 1; x >= 0; x-- {
		if !ascii.IsSpace(token[x]) {
			break
		}
	}
	if x < 0 {
		token = token[:0]
	} else {
		token = token[:x+1]
	}

	return token, d
}

// Remaining return the copy of un-parsed content.
func (bp *Parser) Remaining() []byte {
	return bytes.Clone(bp.content[bp.x:])
}

// RemoveDelimiters remove delimiters delims from current delimiters.
func (bp *Parser) RemoveDelimiters(delims []byte) {
	var (
		newDelims = make([]byte, 0, len(bp.delims))

		oldd  byte
		remd  byte
		found bool
	)
	for _, oldd = range bp.delims {
		found = false
		for _, remd = range delims {
			if remd == oldd {
				found = true
				break
			}
		}
		if !found {
			newDelims = append(newDelims, oldd)
		}
	}
	bp.delims = newDelims
}

// Reset the Parser by setting all internal state to new content and
// delimiters.
func (bp *Parser) Reset(content, delims []byte) {
	bp.content = content
	bp.delims = delims
	bp.x = 0
	bp.size = len(content)
}

// SetDelimiters replace the current delimiters with delims.
func (bp *Parser) SetDelimiters(delims []byte) {
	bp.delims = delims
}

// Skip skip parsing token until one of the delimiters found or
// end-of-content.
func (bp *Parser) Skip() (c byte) {
	var d byte
	for bp.x < bp.size {
		c = bp.content[bp.x]
		for _, d = range bp.delims {
			if c == d {
				bp.x++
				return c
			}
		}
		bp.x++
	}
	return 0
}

// SkipN skip exactly N characters ignoring delimiters.
// It will return the next character after N or 0 if it reach end-of-content.
func (bp *Parser) SkipN(n int) (c byte) {
	var count int
	for bp.x < bp.size {
		c = bp.content[bp.x]
		if count >= n {
			return c
		}
		count++
		bp.x++
	}
	return 0
}

// SkipHorizontalSpaces skip space (" "), tab ("\t"), carriage return
// ("\r"), and form feed ("\f") characters; and return the number of space
// skipped and first non-space character or 0 if it reach end-of-content.
func (bp *Parser) SkipHorizontalSpaces() (n int, c byte) {
	for ; bp.x < bp.size; bp.x++ {
		c = bp.content[bp.x]
		if c == ' ' || c == '\t' || c == '\r' || c == '\f' {
			n++
			continue
		}
		return n, c
	}
	return n, 0
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

// SkipSpaces skip all spaces character (' ', '\f', '\n', '\r', '\t') and
// return the number of spaces skipped and first non-space character or 0 if
// it reach end-of-content.
func (bp *Parser) SkipSpaces() (n int, c byte) {
	for ; bp.x < bp.size; bp.x++ {
		c = bp.content[bp.x]
		if ascii.IsSpace(c) {
			n++
			continue
		}
		return n, c
	}
	return n, 0
}

// Stop the parser, return the remaining unparsed content and its last
// position, and then call Reset to reset the internal state back to zero.
func (bp *Parser) Stop() (remain []byte, pos int) {
	remain = bytes.Clone(bp.content[bp.x:])
	pos = bp.x
	bp.Reset(nil, nil)
	return remain, pos
}

// UnreadN unread N characters and return the character its pointed
// to.
// If N greater than current position index, it will reset the read pointer
// index back to zero.
func (bp *Parser) UnreadN(n int) byte {
	if n > bp.x {
		bp.x = 0
	} else {
		bp.x -= n
	}
	return bp.content[bp.x]
}
