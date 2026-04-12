// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

// Package diff implement text comparison.
package diff

import (
	"bytes"
	"fmt"
	"strings"
)

// Data stores additions, deletions, and changes between two texts.
type Data struct {
	// OldName and NewName stores the file names being compared from
	// calling the [Files].
	OldName string
	NewName string

	oldlines []Line
	newlines []Line

	Adds       []Line
	Dels       []Line
	LineChunks []LineChunk

	// Unified contains the result of [Unified] diff, without the label
	// ("@@ ... @@") and context lines.
	// To print the result of unified diff with context lines use
	// [Data.WriteUnified].
	Unified []Line

	IsMatched bool
}

// Files returns the differences between oldfile and newfile.
func Files(oldfile, newfile string, level int) (diff *Data, err error) {
	logp := `Files`
	oldlines, err := ReadLines(oldfile)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	newlines, err := ReadLines(newfile)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	diff = Lines(oldlines, newlines, level)
	diff.OldName = oldfile
	diff.NewName = newfile
	return diff, nil
}

// Lines returns the difference between old and new lines.
func Lines(oldlines, newlines []Line, level int) (diff *Data) {
	diff = &Data{}
	diff.oldlines = oldlines
	diff.newlines = newlines
	oldlen := len(oldlines)
	newlen := len(newlines)
	x := 0 // cursor for old.
	y := 0 // cursor for new.

	for x < oldlen {
		if y == newlen {
			// New text has been fully examined.
			// The rest is the old text.
			// That means deletion at the end of text.
			diff.PushDel(oldlines[x])
			x++
			continue
		}

		// Compare old line with new line.
		if bytes.Equal(oldlines[x].Val, newlines[y].Val) {
			oldlines[x].NumOther = newlines[y].Num
			newlines[y].NumOther = oldlines[x].Num
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
			diff.PushChange(oldlines[x], newlines[y])
			x++
			y++
			continue
		}

		// Old is empty or contain only whitespaces.
		if oldtrimlen == 0 {
			diff.PushDel(oldlines[x])
			x++
			continue
		}

		// New is empty or contain only whitespaces.
		if newtrimlen == 0 {
			diff.PushAdd(newlines[y])
			y++
			continue
		}

		ratio, _, _ := BytesRatio(oldlines[x].Val, newlines[y].Val,
			DefMatchLen)

		if ratio > DefMatchRatio {
			// Ratio of similar bytes is higher than minimum
			// expectation. So, it must be changes
			diff.PushChange(oldlines[x], newlines[y])
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
			diff.PushChange(oldlines[x], newlines[y])
			x++
			y++
			continue
		}

		// x still missing, means deletion in old text.
		if !foundx && foundy {
			for ; x < yatx && x < oldlen; x++ {
				diff.PushDel(oldlines[x])
			}
			continue
		}

		// we found x but y is missing, its mean addition in new text.
		if foundx && !foundy {
			for ; y < xaty && y < newlen; y++ {
				diff.PushAdd(newlines[y])
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
					diff.PushAdd(newlines[y])
				}

			case addlen == dellen:
				// Both changes occur between lines
				for x < yatx && y < xaty {
					diff.PushChange(oldlines[x], newlines[y])
					x++
					y++
				}
			default:
				for ; x < yatx && x < oldlen; x++ {
					diff.PushDel(oldlines[x])
				}
			}
			continue
		}
	}

	// Check if there is a left over from new text.
	for ; y < newlen; y++ {
		diff.PushAdd(newlines[y])
	}

	if level == LevelWords {
		// Process each changes to find modified chunkes.
		for x, change := range diff.LineChunks {
			adds, dels := Bytes(change.Old.Val, change.New.Val, 0, 0)
			diff.LineChunks[x].Adds = adds
			diff.LineChunks[x].Dels = dels
		}
	}

	diff.checkIsMatched()
	diff.OldName = `oldlines`
	diff.NewName = `newlines`
	return diff
}

// Text returns the difference between old and new text.
func Text(old, new []byte, level int) (diff *Data) {
	oldlines := ParseLines(old)
	newlines := ParseLines(new)
	diff = Lines(oldlines, newlines, level)
	diff.OldName = `old`
	diff.NewName = `new`
	return diff
}

// checkIsMatched set the IsMatched to true if no changes found.
func (diff *Data) checkIsMatched() {
	if len(diff.Adds) != 0 {
		return
	}
	if len(diff.Dels) != 0 {
		return
	}
	if len(diff.LineChunks) != 0 {
		return
	}
	diff.IsMatched = true
}

// PushAdd adds new line to slice of Adds.
func (diff *Data) PushAdd(new Line) {
	new.Kind = LineKindAdd
	diff.Adds = append(diff.Adds, new)
}

// PushChange adds the old and new line to Dels, Adds, and [Data.LineChunks]
// respectively.
func (diff *Data) PushChange(old, new Line) {
	if old.Num != new.Num {
		old.NumOther, new.NumOther = new.Num, old.Num
	}
	change := LineChunk{
		Old: old,
		New: new,
	}
	diff.LineChunks = append(diff.LineChunks, change)
	diff.PushDel(old)
	diff.PushAdd(new)
}

// PushDel adds deletion line to slice of Dels.
func (diff *Data) PushDel(old Line) {
	old.Kind = LineKindDel
	diff.Dels = append(diff.Dels, old)
}

// String returns the additions and deletions with custom format.
// The custom format as follow,
//
//	--- <OldName>
//	+++ <NewName>
//	----
//	<NumOther>/ <Num>: <Kind><QuotedVal>
//	++++
//	<NumOther>/ <Num>: <Kind><Val>
//
// The lines after "----" is the deleted lines from the old text.
// The lines after "++++" is the added lines from the new text.
//
// The NumOther is the line number from the other text.
// The Num is the line number where line is deleted in old text, or added in
// new text.
//
// The Kind is '-' for deletion or '+' for addition.
//
// The QuotedVal value is the line text printed inside double quotes.
func (diff Data) String() (s string) {
	if diff.IsMatched {
		return ``
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n", diff.OldName)
	fmt.Fprintf(&sb, "+++ %s\n", diff.NewName)
	if len(diff.Dels) > 0 {
		sb.WriteString("----\n")
		for _, line := range diff.Dels {
			fmt.Fprintln(&sb, line.String())
		}
	}
	if len(diff.Adds) > 0 {
		sb.WriteString("++++\n")
		for _, line := range diff.Adds {
			fmt.Fprintln(&sb, line.String())
		}
	}
	return sb.String()
}

// findLine returns true if line is found in the beginning of text start at
// line `startat`.
// It also return line number of matching line.
// If no match found, it will return false and `startat` value.
func findLine(line Line, text []Line, startat int) (
	found bool,
	n int,
) {
	textlen := len(text)

	for n = startat; n < textlen; n++ {
		if bytes.Equal(line.Val, text[n].Val) {
			return true, n
		}
	}

	return false, startat
}
