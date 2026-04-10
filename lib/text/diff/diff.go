// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

// Package diff implement text comparison.
package diff

import (
	"bytes"
	"fmt"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
)

var (
	// DefDelimiter define default delimiter for new line.
	DefDelimiter = byte('\n')
)

// Data stores additions, deletions, and changes between two texts.
type Data struct {
	Adds    []Line
	Dels    []Line
	Changes LineChanges

	IsMatched bool
}

// Text search the difference between two texts.
func Text(before, after []byte, level int) (diffs Data) {
	beforeLines := ParseLines(before)
	afterLines := ParseLines(after)
	return Lines(beforeLines, afterLines, level)
}

// Lines search the difference between two Lines.
func Lines(oldlines, newlines []Line, level int) (diffs Data) {
	oldlen := len(oldlines)
	newlen := len(newlines)
	x := 0
	y := 0

	for x < oldlen {
		if y == newlen {
			// New text has been fully examined.
			// The rest is the old text.
			// That means deletion at the end of text.
			diffs.PushDel(oldlines[x])
			x++
			continue
		}

		// Compare old line with new line.
		if IsEqual(oldlines[x].Val, newlines[y].Val) {
			x++
			y++
			continue
		}

		// Check for whitespace changes
		oldlinetrim := bytes.TrimSpace(oldlines[x].Val)
		newlinetrim := bytes.TrimSpace(newlines[y].Val)
		oldtrimlen := len(oldlinetrim)
		newtrimlen := len(newlinetrim)

		// Both are empty, probably one of them is changing
		if oldtrimlen == 0 && newtrimlen == 0 {
			diffs.PushChange(oldlines[x], newlines[y])
			x++
			y++
			continue
		}

		// Old is empty or contain only whitespaces.
		if oldtrimlen <= 0 {
			diffs.PushDel(oldlines[x])
			x++
			continue
		}

		// New is empty or contain only whitespaces.
		if newtrimlen <= 0 {
			diffs.PushAdd(newlines[y])
			y++
			continue
		}

		ratio, _, _ := BytesRatio(oldlines[x].Val, newlines[y].Val,
			DefMatchLen)

		if ratio > DefMatchRatio {
			// Ratio of similar bytes is higher than minimum
			// expectation. So, it must be changes
			diffs.PushChange(oldlines[x], newlines[y])
			x++
			y++
			continue
		}

		// x is not equal with y, search down...
		foundx, xaty := findLine(oldlines[x], newlines, y+1)

		// Cross check the y with the rest of x...
		foundy, yatx := findLine(newlines[y], oldlines, x+1)

		// Both line is missing, its mean changes on current line
		if !foundx && !foundy {
			diffs.PushChange(oldlines[x], newlines[y])
			x++
			y++
			continue
		}

		// x still missing, means deletion in old text.
		if !foundx && foundy {
			for ; x < yatx && x < oldlen; x++ {
				diffs.PushDel(oldlines[x])
			}
			continue
		}

		// we found x but y is missing, its mean addition in new text.
		if foundx && !foundy {
			for ; y < xaty && y < newlen; y++ {
				diffs.PushAdd(newlines[y])
			}
			continue
		}

		if foundx && foundy {
			// We found x and y. Check which one is the
			// addition or deletion based on line range.
			addlen := xaty - y
			dellen := yatx - x

			switch {
			case addlen < dellen:
				for ; y < xaty && y < newlen; y++ {
					diffs.PushAdd(newlines[y])
				}

			case addlen == dellen:
				// Both changes occur between lines
				for x < yatx && y < xaty {
					diffs.PushChange(oldlines[x],
						newlines[y])
					x++
					y++
				}
			default:
				for ; x < yatx && x < oldlen; x++ {
					diffs.PushDel(oldlines[x])
				}
			}
			continue
		}
	}

	// Check if there is a left over from new text.
	for ; y < newlen; y++ {
		diffs.PushAdd(newlines[y])
	}

	if level == LevelWords {
		// Process each changes to find modified chunkes.
		for x, change := range diffs.Changes {
			adds, dels := Bytes(change.Old.Val, change.New.Val, 0, 0)
			diffs.Changes[x].Adds = adds
			diffs.Changes[x].Dels = dels
		}
	}

	diffs.checkIsMatched()

	return diffs
}

// checkIsMatched set the IsMatched to true if no changes found.
func (diffs *Data) checkIsMatched() {
	if len(diffs.Adds) != 0 {
		return
	}
	if len(diffs.Dels) != 0 {
		return
	}
	if len(diffs.Changes) != 0 {
		return
	}
	diffs.IsMatched = true
}

// PushAdd will add new line to diff set.
func (diffs *Data) PushAdd(new Line) {
	diffs.Adds = append(diffs.Adds, new)
}

// PushDel will add deletion line to diff set.
func (diffs *Data) PushDel(old Line) {
	diffs.Dels = append(diffs.Dels, old)
}

// PushChange set to diff data.
func (diffs *Data) PushChange(old, new Line) {
	change := NewLineChange(old, new)

	diffs.Changes = append(diffs.Changes, *change)
}

// GetAllAdds return chunks of additions including in line changes.
func (diffs *Data) GetAllAdds() (chunks text.Chunks) {
	for _, add := range diffs.Adds {
		chunks = append(chunks, text.Chunk{StartAt: 0, V: add.Val})
	}
	chunks = append(chunks, diffs.Changes.GetAllAdds()...)
	return
}

// GetAllDels return chunks of deletions including in line changes.
func (diffs *Data) GetAllDels() (chunks text.Chunks) {
	for _, del := range diffs.Dels {
		chunks = append(chunks, text.Chunk{StartAt: 0, V: del.Val})
	}
	chunks = append(chunks, diffs.Changes.GetAllDels()...)
	return
}

// String return formatted data.
func (diffs Data) String() (s string) {
	var sb strings.Builder

	if len(diffs.Dels) > 0 {
		sb.WriteString("----\n")
		for _, line := range diffs.Dels {
			fmt.Fprintf(&sb, "%d - %q\n", line.Num, line.Val)
		}
	}

	if len(diffs.Adds) > 0 {
		sb.WriteString("++++\n")
		for _, line := range diffs.Adds {
			fmt.Fprintf(&sb, "%d + %q\n", line.Num, line.Val)
		}
	}

	if len(diffs.Changes) > 0 {
		sb.WriteString("--++\n")
		for _, change := range diffs.Changes {
			sb.WriteString(change.String())
		}
	}

	return sb.String()
}
