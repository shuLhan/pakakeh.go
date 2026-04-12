// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package diff

import (
	inbytes "git.sr.ht/~shulhan/pakakeh.go/internal/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
)

// Bytes returns the character differences (additions and deletion) between
// old and new bytes, start at specific position.
func Bytes(old, new []byte, oldat, newat int) (adds, dels text.Chunks) {
	oldlen := len(old)
	newlen := len(new)

	minlen := 0
	minlen = min(oldlen, newlen)

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
			adds = append(adds, text.Chunk{StartAt: oldat + y, V: v})
		} else {
			v := old[x:]
			dels = append(dels, text.Chunk{StartAt: oldat + x, V: v})
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
		adds = append(adds, text.Chunk{StartAt: newat + y, V: v})
		return adds, dels
	}

	// Case 3: deletion in old line.
	if y == yend+1 {
		v := old[x : xend+1]
		dels = append(dels, text.Chunk{StartAt: oldat + x, V: v})
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
	minlen = min(oldleftlen, DefMatchLen)
	xtoken := oldleft[:minlen]

	xaty := inbytes.TokenFind(newleft, xtoken, 0)

	// Get miniminal token to search in the old left over.
	minlen = min(newleftlen, DefMatchLen)
	ytoken := newleft[:minlen]

	yatx := inbytes.TokenFind(oldleft, ytoken, 0)

	// Case 4:
	// We did not find matching token of x in y, its mean the some chunk
	// in x and y has been replaced.
	if xaty < 0 && yatx < 0 {
		addsleft, delsleft := searchForward(oldat, newat, &x, &y, &oldleft,
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
					StartAt: oldat + x,
					V:       newleft,
				})
			}
			return adds, dels
		}
		if len(newleft) == 0 {
			if len(oldleft) > 0 {
				dels = append(dels, text.Chunk{
					StartAt: newat + y,
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
		adds = append(adds, text.Chunk{StartAt: newat + y, V: v})
		newleft = new[y+xaty : yend+1]
	} else if yatx >= 0 {
		// Case 3: We found y token at x: yatx. Previous byte before that must
		// be a deletion.
		v := old[x : x+yatx]
		dels = append(dels, text.Chunk{StartAt: oldat + x, V: v})
		oldleft = old[x+yatx : xend+1]
	}

	addsleft, delsleft := Bytes(oldleft, newleft, oldat+x, newat+y)

	if len(addsleft) > 0 {
		adds = append(adds, addsleft...)
	}
	if len(delsleft) > 0 {
		dels = append(dels, delsleft...)
	}

	return adds, dels
}

// BytesRatio returns the ratio of matching bytes between old and new.
// The ratio in in range of 0.0 to 1.0, where 1.0 when both are similar, and
// 0.0 when no single character matchs found.
// The minTokenLen define the minimum length of token for searching in both of
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
		minlen = min(oldlen, minTokenLen)

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

		minlen = min(newlen, oldlen)

		x, y = 0, 0
		// start again from beginning...
	}

	ratio = float32(m) / float32(maxlen)

	return ratio, m, maxlen
}

func searchForward(oldat, newat int, x, y *int, oldleft, newleft *[]byte) (
	adds, dels text.Chunks,
) {
	oldleftlen := len(*oldleft)
	newleftlen := len(*newleft)

	minlen := min(oldleftlen, DefMatchLen)

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

	minlen = min(newleftlen, DefMatchLen)

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
		dels = append(dels, text.Chunk{StartAt: oldat + *x, V: *oldleft})
		adds = append(adds, text.Chunk{StartAt: newat + *y, V: *newleft})
		*oldleft = []byte{}
		*newleft = []byte{}
		return adds, dels
	}

	// Some chunk has been replaced.
	v := (*oldleft)[:xx]
	dels = append(dels, text.Chunk{StartAt: oldat + *x, V: v})
	*oldleft = (*oldleft)[xx:]
	*x += xx

	v = (*newleft)[:yy]
	adds = append(adds, text.Chunk{StartAt: newat + *y, V: v})
	*newleft = (*newleft)[yy:]
	*y += yy

	return adds, dels
}
