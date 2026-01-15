// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package strings

import (
	"fmt"
	"os"

	libascii "git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

// Parser implement text parsing over string.
type Parser struct {
	file   string
	delims string
	v      string // v contains the text to be parsed.
	token  []rune // token that has been parsed.
	x      int    // x is the position of read in v.
	d      rune   // d is one of delims character that terminated parsing.
}

// LinesOfFile parse the content of file and return non-empty lines.
func LinesOfFile(file string) ([]string, error) {
	p, err := OpenForParser(file, ``)
	if err != nil {
		return nil, fmt.Errorf(`LinesOfFile: %w`, err)
	}
	return p.Lines(), nil
}

// NewParser create and initialize parser from content and delimiters.
func NewParser(content, delims string) (p *Parser) {
	p = &Parser{
		token: make([]rune, 0, 16),
	}

	p.Load(content, delims)

	return p
}

// OpenForParser create and initialize the Parser using content from file.
// If delimiters is empty, it would default to all whitespaces characters.
func OpenForParser(file, delims string) (p *Parser, err error) {
	v, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	p = NewParser(string(v), delims)
	p.file = file

	return p, nil
}

// AddDelimiters append new delimiter to existing parser.
func (p *Parser) AddDelimiters(delims string) {
	var found bool
	for _, newd := range delims {
		found = false
		for _, oldd := range p.delims {
			if oldd == newd {
				found = true
				break
			}
		}
		if !found {
			p.delims += string(newd)
		}
	}
}

// Close the parser by resetting all its internal state to zero value.
func (p *Parser) Close() {
	p.file = ``
	p.delims = ``
	p.x = 0
	p.v = ``
	p.token = p.token[:0]
	p.d = 0
}

// isDelim true if r is one of delimiters.
func (p *Parser) isDelim(r rune) bool {
	var d rune
	for _, d = range p.delims {
		if r == d {
			return true
		}
	}
	return false
}

// Lines return all non-empty lines from the content.
func (p *Parser) Lines() []string {
	var start, end int

	lines := make([]string, 0)

	for x := p.x; x < len(p.v); x++ {
		// Skip white spaces on beginning ...
		for ; x < len(p.v); x++ {
			if p.v[x] == ' ' || p.v[x] == '\t' || p.v[x] == '\r' || p.v[x] == '\f' {
				continue
			}
			break
		}
		start = x
		for ; x < len(p.v); x++ {
			if p.v[x] != '\n' {
				continue
			}
			break
		}

		// Skip white spaces at the end ...
		for end = x - 1; end > start; end-- {
			if p.v[end] == ' ' || p.v[end] == '\t' ||
				p.v[end] == '\r' || p.v[end] == '\f' {
				continue
			}
			break
		}
		end++
		if start == end {
			// Skip empty lines
			continue
		}

		line := p.v[start:end]
		lines = append(lines, line)
	}

	p.x = len(p.v)

	return lines
}

// Load the new content and delimiters.
func (p *Parser) Load(content, delims string) {
	p.Close()
	p.v = content
	if len(delims) == 0 {
		p.delims = string(libascii.Spaces)
	} else {
		p.delims = delims
	}
}

// Line read and return a single line.
// On success it will return a string without '\n' and new line character.
// In case of EOF it will return the last line and 0.
func (p *Parser) Line() (string, rune) {
	p.d = 0
	p.token = p.token[:0]

	for x, r := range p.v[p.x:] {
		if r == '\n' {
			p.d = r
			p.x += x + 1
			return string(p.token), p.d
		}
		p.token = append(p.token, r)
	}
	p.x = len(p.v)
	return string(p.token), 0
}

// SetDelimiters replace the current delimiters with delims.
func (p *Parser) SetDelimiters(delims string) {
	p.delims = delims
}

// Stop the parser, return the remaining unparsed content and its last
// position, and then call Close to reset the internal state back to zero.
func (p *Parser) Stop() (remain string, pos int) {
	pos = p.x
	remain = p.v[pos:]
	p.Close()
	return remain, pos
}

// Read read the next token from content until one of the delimiter found.
// if no delimiter found, its mean all of content has been read, the returned
// delimiter will be 0.
func (p *Parser) Read() (string, rune) {
	p.d = 0
	p.token = p.token[:0]

	if p.x >= len(p.v) {
		return ``, 0
	}

	for x, r := range p.v[p.x:] {
		for _, d := range p.delims {
			if r == d {
				p.d = d
				p.x += x + 1
				return string(p.token), p.d
			}
		}

		p.token = append(p.token, r)
	}

	p.x = len(p.v)
	return string(p.token), 0
}

// ReadEscaped read the next token from content until one of the delimiter
// found, unless its escaped with value of esc character.
//
// For example, if the content is "a b" and one of the delimiter is " ",
// escaping it with "\" will return as "a b" not "a".
func (p *Parser) ReadEscaped(esc rune) (string, rune) {
	var isEscaped bool

	p.token = p.token[:0]

	if p.x >= len(p.v) {
		p.d = 0
		return ``, 0
	}

	for x, r := range p.v[p.x:] {
		if r == esc {
			if isEscaped {
				p.token = append(p.token, r)
				isEscaped = false
				continue
			}
			isEscaped = true
			continue
		}
		for _, d := range p.delims {
			if r == d {
				if isEscaped {
					isEscaped = false
					break
				}

				p.d = d
				p.x += x + 1
				return string(p.token), p.d
			}
		}

		p.token = append(p.token, r)
	}

	p.d = 0
	p.x = len(p.v)
	return string(p.token), p.d
}

// ReadNoSpace read the next token until one of the delimiter found, with
// leading and trailing spaces are ignored.
func (p *Parser) ReadNoSpace() (v string, r rune) {
	p.d = 0
	p.token = p.token[:0]

	if p.x >= len(p.v) {
		return ``, 0
	}

	var x int

	// Skip leading spaces.
	for x, r = range p.v[p.x:] {
		if isHorizontalSpace(r) {
			continue
		}
		break
	}
	p.x += x

	for x, r = range p.v[p.x:] {
		if p.isDelim(r) {
			p.d = r
			break
		}
		p.token = append(p.token, r)
	}

	p.x += x + 1 // +1 to skip the delimiter.

	// Remove trailing spaces.
	for x = len(p.token) - 1; x >= 0; x-- {
		if isHorizontalSpace(p.token[x]) {
			continue
		}
		break
	}
	if x < 0 {
		// Empty token.
		return ``, p.d
	}
	p.token = p.token[:x+1]

	return string(p.token), p.d
}

// ReadEnclosed read the token inside opening and closing characters, ignoring
// all delimiters that previously set.
//
// It will return the parsed token and closed character if closed character
// found, otherwise it will token with 0.
func (p *Parser) ReadEnclosed(open, closed rune) (string, rune) {
	for x, r := range p.v[p.x:] {
		if x == 0 {
			if r == open {
				continue
			}
		}
		if r == closed {
			p.d = closed
			p.x += x + 1
			return string(p.token), p.d
		}

		p.token = append(p.token, r)
	}

	p.d = 0
	p.x = len(p.v)
	return p.v, 0
}

// RemoveDelimiters from current parser.
func (p *Parser) RemoveDelimiters(dels string) {
	var (
		newdelims string
		found     bool
	)

	for _, oldd := range p.delims {
		found = false
		for _, r := range dels {
			if r == oldd {
				found = true
				break
			}
		}
		if !found {
			newdelims += string(oldd)
		}
	}

	p.delims = newdelims
}

// Skip parsing n characters or EOF if n is greater then length of content.
func (p *Parser) Skip(n int) {
	if p.x+n >= len(p.v) {
		p.x = len(p.v)
		p.d = 0
	} else {
		p.x += n
	}
}

// SkipHorizontalSpaces skip all space (" "), tab ("\t"), carriage return
// ("\r"), and form feed ("\f") characters; and return the first character
// found, probably new line.
func (p *Parser) SkipHorizontalSpaces() rune {
	for x, r := range p.v[p.x:] {
		switch r {
		case ' ', '\t', '\r', '\f':
		default:
			p.x += x
			p.d = r
			return r
		}
	}

	p.d = 0
	p.x = len(p.v)

	return 0
}

// SkipLine skip all characters until new line.
// It will return the first character after new line or 0 if EOF.
func (p *Parser) SkipLine() rune {
	for x, r := range p.v[p.x:] {
		if r == '\n' {
			p.x += x + 1
			if p.x >= len(p.v) {
				p.d = 0
			} else {
				p.d = r
			}
			return p.d
		}
	}

	// All contents has been read, no new line found.
	p.x = len(p.v)
	p.d = 0

	return 0
}

// isHorizontalSpace true if r is space, tab, carriage return, or form feed.
func isHorizontalSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\f'
}
