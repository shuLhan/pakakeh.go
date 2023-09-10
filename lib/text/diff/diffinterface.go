// Copyright 2018 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"bufio"
	"io"
	"os"

	inbytes "github.com/shuLhan/share/internal/bytes"
	"github.com/shuLhan/share/lib/text"
)

const (
	// LevelLines define that we want only lines change set.
	LevelLines = iota
	// LevelWords define that we want the change not only capture the
	// different per line, but also changes inside the line.
	LevelWords
)

const (
	// DefMatchLen minimum number of bytes used for searching the next
	// matched chunk in line.
	DefMatchLen = 5
	// DefMatchRatio define default minimum match ratio to be considered as
	// change.
	DefMatchRatio = 0.7
)

// ReadLines return lines in the file `f`.
func ReadLines(f string) (lines text.Lines, e error) {
	fd, e := os.Open(f)

	if e != nil {
		return
	}

	reader := bufio.NewReader(fd)

	n := 1
	for {
		line, e := reader.ReadBytes(DefDelimiter)

		if e != nil {
			if e == io.EOF {
				break
			}
			return lines, e
		}

		lines = append(lines, text.Line{N: n, V: line})
		n++
	}

	e = fd.Close()

	return lines, e
}

// IsEqual compare two slice of bytes and return true if equal or false
// otherwise.
func IsEqual(oldb, newb []byte) (equal bool) {
	oldblen := len(oldb)
	newblen := len(newb)

	// Do not compare the length, because we care about the index.

	minlen := 0
	switch {
	case oldblen < newblen:
		minlen = oldblen
	case oldblen == newblen:
		minlen = oldblen
	default:
		minlen = newblen
	}

	at := 0
	for ; at < minlen; at++ {
		if oldb[at] != newb[at] {
			return
		}
	}

	return oldblen == newblen
}

// BytesRatio compare two slice of bytes and return ratio of matching bytes.
// The ratio in in range of 0.0 to 1.0, where 1.0 if both are similar, and 0.0
// if no matchs even found.
// `minTokenLen` define the minimum length of token for searching in both of
// slice.
func BytesRatio(old, newline []byte, minTokenLen int) (ratio float32, m int, maxlen int) {
	x, y := 0, 0

	oldlen := len(old)
	newlen := len(newline)
	minlen := oldlen
	maxlen = newlen
	if newlen < oldlen {
		minlen = newlen
		maxlen = oldlen
	}

	if minTokenLen < 0 {
		minTokenLen = DefMatchLen
	}

	for {
		// Count matching bytes from beginning of slice.
		for x < minlen {
			if old[x] != newline[y] {
				break
			}
			m++
			x++
			y++
		}

		if x == minlen {
			// All bytes is matched but probably some trailing in
			// one of them.
			break
		}

		// Count matching bytes from end of slice
		xend := oldlen - 1
		yend := newlen - 1

		for xend >= x && yend >= y {
			if old[xend] != newline[yend] {
				break
			}
			m++
			xend--
			yend--
		}

		// One of the line have changes in the middle.
		if xend == x || yend == y {
			break
		}

		// Cut the matching bytes
		old = old[x : xend+1]
		newline = newline[y : yend+1]
		oldlen = len(old)

		// Get minimal token to search in the newline left over.
		minlen = minTokenLen
		if oldlen < minlen {
			minlen = oldlen
		}

		// Search old token in newline, chunk by chunk.
		x = 0
		y = -1
		max := oldlen - minlen
		for ; x < max; x++ {
			token := old[x : x+minlen]

			y = inbytes.TokenFind(newline, token, 0)
			if y > 0 {
				break
			}
		}

		if y < 0 {
			// We did not found anything.
			break
		}

		// Cut the changes
		old = old[x:]
		newline = newline[y:]
		oldlen = len(old)
		newlen = len(newline)

		minlen = oldlen
		if newlen < minlen {
			minlen = newlen
		}

		x, y = 0, 0
		// start again from beginning...
	}

	ratio = float32(m) / float32(maxlen)

	return ratio, m, maxlen
}

// findLine return true if line is found in text beginning at line `startat`.
// It also return line number of matching line.
// If no match found, it will return false and `startat` value.
func findLine(line text.Line, text text.Lines, startat int) (
	found bool,
	n int,
) {
	textlen := len(text)

	for n = startat; n < textlen; n++ {
		if IsEqual(line.V, text[n].V) {
			return true, n
		}
	}

	return false, startat
}

// Files compare two files.
func Files(oldf, newf string, level int) (diffs Data, e error) {
	oldlines, e := ReadLines(oldf)
	if e != nil {
		return
	}
	newlines, e := ReadLines(newf)
	if e != nil {
		return
	}
	diffs = Lines(oldlines, newlines, level)
	return diffs, nil
}

// Bytes given two similar lines, find and return the differences (additions and
// deletion) between them.
//
// Case 1: addition on new or deletion on old.
//
//	old: 00000
//	new: 00000111
//
// or
//
//	old: 00000111
//	new: 00000
//
// Case 2: addition on new line
//
//	old: 000000
//	new: 0001000
//
// Case 3: deletion on old line (reverse of case 2)
//
//	old: 0001000
//	new: 000000
//
// Case 4: change happened in the beginning
//
//	old: 11000
//	new: 22000
//
// Case 5: both changed
//
//	old: 0001000
//	new: 0002000
func Bytes(old, new []byte, atx, aty int) (adds, dels text.Chunks) {
	oldlen := len(old)
	newlen := len(new)

	minlen := 0
	if oldlen < newlen {
		minlen = oldlen
	} else {
		minlen = newlen
	}

	// Find the position of unmatched byte from the beginning.
	x, y := 0, 0
	for ; x < minlen; x++ {
		if old[x] != new[x] {
			break
		}
	}
	y = x

	// Case 1: Check if addition or deletion is at the end.
	if x == minlen {
		if oldlen < newlen {
			v := new[y:]
			adds = append(adds, text.Chunk{StartAt: atx + y, V: v})
		} else {
			v := old[x:]
			dels = append(dels, text.Chunk{StartAt: atx + x, V: v})
		}
		return adds, dels
	}

	// Find the position of unmatched byte from the end
	xend := oldlen - 1
	yend := newlen - 1

	for xend >= x && yend >= y {
		if old[xend] != new[yend] {
			break
		}
		xend--
		yend--
	}

	// Case 2: addition in new line.
	if x == xend+1 {
		v := new[y : yend+1]
		adds = append(adds, text.Chunk{StartAt: aty + y, V: v})
		return adds, dels
	}

	// Case 3: deletion in old line.
	if y == yend+1 {
		v := old[x : xend+1]
		dels = append(dels, text.Chunk{StartAt: atx + x, V: v})
		return adds, dels
	}

	// Calculate possible match len.
	// After we found similar bytes in the beginning and end of line, now
	// we have `n` number of bytes left in old and new.
	oldleft := old[x : xend+1]
	newleft := new[y : yend+1]
	oldleftlen := len(oldleft)
	newleftlen := len(newleft)

	// Get minimal token to search in the new left over.
	minlen = DefMatchLen
	if oldleftlen < DefMatchLen {
		minlen = oldleftlen
	}
	xtoken := oldleft[:minlen]

	xaty := inbytes.TokenFind(newleft, xtoken, 0)

	// Get miniminal token to search in the old left over.
	minlen = DefMatchLen
	if newleftlen < DefMatchLen {
		minlen = newleftlen
	}
	ytoken := newleft[:minlen]

	yatx := inbytes.TokenFind(oldleft, ytoken, 0)

	// Case 4:
	// We did not find matching token of x in y, its mean the some chunk
	// in x and y has been replaced.
	if xaty < 0 && yatx < 0 {
		addsleft, delsleft := searchForward(atx, aty, &x, &y, &oldleft,
			&newleft)

		if len(addsleft) > 0 {
			adds = append(adds, addsleft...)
		}
		if len(delsleft) > 0 {
			dels = append(dels, delsleft...)
		}

		// Check for possible empty left
		if len(oldleft) == 0 {
			if len(newleft) > 0 {
				adds = append(adds, text.Chunk{
					StartAt: atx + x,
					V:       newleft,
				})
			}
			return adds, dels
		}
		if len(newleft) == 0 {
			if len(oldleft) > 0 {
				dels = append(dels, text.Chunk{
					StartAt: aty + y,
					V:       oldleft,
				})
			}
			return adds, dels
		}
	}

	// Case 5: is combination of case 2 and 3.
	// Case 2: We found x token at y: xaty. Previous byte before that must
	// be an addition.
	if xaty >= 0 {
		v := new[y : y+xaty]
		adds = append(adds, text.Chunk{StartAt: aty + y, V: v})
		newleft = new[y+xaty : yend+1]
	} else if yatx >= 0 {
		// Case 3: We found y token at x: yatx. Previous byte before that must
		// be a deletion.
		v := old[x : x+yatx]
		dels = append(dels, text.Chunk{StartAt: atx + x, V: v})
		oldleft = old[x+yatx : xend+1]
	}

	addsleft, delsleft := Bytes(oldleft, newleft, atx+x, aty+y)

	if len(addsleft) > 0 {
		adds = append(adds, addsleft...)
	}
	if len(delsleft) > 0 {
		dels = append(dels, delsleft...)
	}

	return adds, dels
}

func searchForward(atx, aty int, x, y *int, oldleft, newleft *[]byte) (
	adds, dels text.Chunks,
) {
	oldleftlen := len(*oldleft)
	newleftlen := len(*newleft)

	minlen := DefMatchLen
	if oldleftlen < minlen {
		minlen = oldleftlen
	}

	// Loop through old line to find matching token
	xaty := -1
	xx := 1
	for ; xx < oldleftlen-minlen; xx++ {
		token := (*oldleft)[xx : xx+minlen]

		xaty = inbytes.TokenFind(*newleft, token, 0)
		if xaty > 0 {
			break
		}
	}

	minlen = DefMatchLen
	if newleftlen < minlen {
		minlen = newleftlen
	}

	yatx := -1
	yy := 1
	for ; yy < newleftlen-minlen; yy++ {
		token := (*newleft)[yy : yy+minlen]

		yatx = inbytes.TokenFind(*oldleft, token, 0)
		if yatx > 0 {
			break
		}
	}

	if xaty < 0 && yatx < 0 {
		// still no token found, means whole chunk has been replaced.
		dels = append(dels, text.Chunk{StartAt: atx + *x, V: *oldleft})
		adds = append(adds, text.Chunk{StartAt: aty + *y, V: *newleft})
		*oldleft = []byte{}
		*newleft = []byte{}
		return adds, dels
	}

	// Some chunk has been replaced.
	v := (*oldleft)[:xx]
	dels = append(dels, text.Chunk{StartAt: atx + *x, V: v})
	*oldleft = (*oldleft)[xx:]
	*x += xx

	v = (*newleft)[:yy]
	adds = append(adds, text.Chunk{StartAt: aty + *y, V: v})
	*newleft = (*newleft)[yy:]
	*y += yy

	return adds, dels
}
