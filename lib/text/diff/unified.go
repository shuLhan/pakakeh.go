// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package diff

import (
	"fmt"
	"io"
	"slices"
)

const defautlNumContext = 3

// Unified returns the unified changes between old and new bytes.
//
// References,
//   - https://www.gnu.org/software/diffutils/manual/html_node/Unified-Format.html
//   - https://www.gnu.org/software/diffutils/manual/html_node/Hunks.html
func Unified(old, new []byte) (diff *Data) {
	diff = Text(old, new, LevelLines)

	// Mergeds the deletions and additions lines.
	// The deleted lines must be before the added lines, so we merge them
	// first.

	diff.Unified = make([]Line, 0, len(diff.Adds)+len(diff.Dels)+(defautlNumContext*2))
	diff.Unified = append(diff.Unified, diff.Dels...)

	var previdx int
	var prev *Line
	adds := diff.Adds
	if len(adds) != 0 {
		prev = &adds[0]
		previdx = findInsertIndex(diff.Unified, prev, 0)
		if previdx == -1 {
			diff.Unified = append(diff.Unified, *prev)
			previdx = len(diff.Unified)
		} else {
			diff.Unified = slices.Insert(diff.Unified, previdx, *prev)
			previdx++
		}
		adds = adds[1:]
	}
	for _, line := range adds {
		if prev.Num+1 == line.Num {
			diff.Unified = slices.Insert(diff.Unified, previdx, line)
			prev = &line
			previdx++
			continue
		}
		previdx = findInsertIndex(diff.Unified, &line, previdx)
		if previdx == -1 {
			diff.Unified = append(diff.Unified, line)
			prev = &line
			previdx = len(diff.Unified)
			continue
		}
		diff.Unified = slices.Insert(diff.Unified, previdx, line)
		prev = &line
		previdx++
	}

	return diff
}

// findInsertIndex returns the slice index to insert the next line into
// unified, beginning from index start.
//
// It will returns -1 if unified is empty or all lines number in unified less
// or equal than next line.
func findInsertIndex(unified []Line, next *Line, start int) (idx int) {
	if len(unified) == 0 {
		return -1
	}
	nextNum := max(next.Num, next.NumOther)
	var prev *Line
	for start < len(unified) {
		line := &unified[start]
		if prev == nil {
			if line.Num <= nextNum {
				prev = line
				start++
				continue
			}
			return start
		}
		if line.Num <= nextNum {
			prev = line
			start++
			continue
		}
		if prev.Num+1 == line.Num {
			prev = line
			start++
			continue
		}
		return start
	}
	return -1
}

// WriteUnified writes the result of unified diff into w with ncontext number
// of unmodified lines before and after it.
//
// The ncontext must be 0 or positive, otherwise it will be set to 0 (no
// context added).
func (diff Data) WriteUnified(w io.Writer, ncontext int) (err error) {
	if len(diff.Unified) == 0 {
		return nil
	}
	if ncontext < 0 {
		ncontext = 0
	}
	var hunks [][]Line
	var hunk []Line
	uni := diff.Unified
	for len(uni) > 0 {
		hunk, uni = diff.hunk(uni, ncontext)
		if len(hunk) == 0 {
			break
		}
		hunk = completeHunk(hunk)
		hunks = append(hunks, hunk)
	}
	return diff.writeHunks(w, hunks, false)
}

// hunk returns combination of unified lines with context lines.
//
// Remember that we use zero index slices with line number start with 1.
// So, assume that current line is at idx 5 (line number 6),
//
//   - for the appendBefore, the end must be line.Num-1 (at idx 5).
//
//   - for the appendAfter, the start must be line.Num (at idx 6).
//
//     idx  lines
//     1
//     2
//     3
//     4    {Num:5 Kind:' ' Val:...}
//     5    {Num:6 Kind:'-' Val:...} <-- end (appendBefore) (2,3,4)
//     6    {Num:7 Kind:' ' Val:...} <-- start (appendAfter) (6,7,8)
func (diff *Data) hunk(unified []Line, nctx int) (hunk []Line, remains []Line) {
	x := 0
	prev := unified[x]
	switch prev.Kind {
	case LineKindAdd:
		hunk = appendBefore(hunk, diff.newlines, prev.Num-1, nctx)
	case LineKindDel:
		hunk = appendBefore(hunk, diff.oldlines, prev.Num-1, nctx)
	}
	hunk = append(hunk, prev)
	x++
	if x == len(unified) {
		switch prev.Kind {
		case LineKindAdd:
			hunk = appendAfter(hunk, diff.newlines, prev.Num, nctx)
		case LineKindDel:
			hunk = appendAfter(hunk, diff.oldlines, prev.Num, nctx)
		}
		return hunk, nil
	}
	for x < len(unified) {
		line := unified[x]
		if prev.Kind == line.Kind {
			if prev.Num+1 == line.Num {
				// Case,
				// 	1: -old
				// 	2: -old
				hunk = append(hunk, line)
				prev = line
				x++
				continue
			}
		}
		count := nctx
		gap := computeGap(hunk, prev, line)
		if gap <= (nctx * 2) {
			count = gap
		}
		// Add n number of lines after.
		switch prev.Kind {
		case LineKindAdd:
			hunk = appendAfter(hunk, diff.newlines, prev.Num, count)
		case LineKindDel:
			hunk = appendAfter(hunk, diff.oldlines, prev.Num, count)
		}
		if gap <= (nctx * 2) {
			hunk = append(hunk, line)
			prev = line
			x++
			continue
		}
		remains = unified[x:]
		return hunk, remains
	}
	// Changes on the end.
	switch prev.Kind {
	case LineKindAdd:
		hunk = appendAfter(hunk, diff.newlines, prev.Num, nctx)
	case LineKindDel:
		hunk = appendAfter(hunk, diff.oldlines, prev.Num, nctx)
	}
	remains = unified[x:]
	return hunk, remains
}

func computeGap(hunk []Line, prev, next Line) (gap int) {
	if prev.Kind == next.Kind {
		return (next.Num - prev.Num) - 1
	}
	// Find the previous line with the same kind as next or unmodified.
	for _, line := range slices.Backward(hunk) {
		if line.Kind == next.Kind || line.Kind == LineKindNone {
			return (next.Num - line.Num) - 1
		}
	}
	return next.Num - 1
}

// appendBefore prepends lines from src at least nctx lines.
func appendBefore(hunk, src []Line, end, nctx int) []Line {
	start := max(0, end-nctx)
	for start < end {
		hunk = append(hunk, src[start])
		start++
	}
	return hunk
}

func appendAfter(hunk, src []Line, start, count int) []Line {
	end := min(len(src), start+count)
	for start < end {
		hunk = append(hunk, src[start])
		start++
	}
	return hunk
}

// completeHunk generates label for hunk "@@ ... @@".
func completeHunk(hunk []Line) []Line {
	firstLine := hunk[0]
	var firstChange *Line
	for _, line := range hunk {
		if line.Kind == LineKindNone {
			continue
		}
		firstChange = &line
		break
	}

	var oldStart, newStart int

	if firstChange.Kind == LineKindAdd {
		oldStart = firstLine.NumOther
		newStart = firstLine.Num
	} else {
		oldStart = firstLine.Num
		newStart = firstLine.NumOther
	}
	var oldCount int
	var newCount int
	for _, line := range hunk {
		switch line.Kind {
		case LineKindAdd:
			newCount++
		case LineKindDel:
			oldCount++
		case LineKindNone:
			oldCount++
			newCount++
		}
	}

	var args = make([]any, 0, 4)
	var oldfmt = `-%d,%d`
	if oldCount <= 1 {
		oldfmt = `-%d`
		args = append(args, oldStart)
	} else {
		args = append(args, oldStart, oldCount)
	}
	var newfmt = `+%d,%d`
	if newCount <= 1 {
		newfmt = `+%d`
		args = append(args, newStart)
	} else {
		args = append(args, newStart, newCount)
	}

	label := fmt.Sprintf(`@@ `+oldfmt+` `+newfmt+` @@`, args...)
	hunkLabel := Line{
		Val:  []byte(label),
		Kind: LineKindLabel,
	}
	return slices.Insert(hunk, 0, hunkLabel)
}

func (diff Data) writeHunks(w io.Writer, hunks [][]Line, withLineNumber bool) (err error) {
	fmt.Fprintln(w, `--- `+diff.OldName)
	fmt.Fprintln(w, `+++ `+diff.NewName)
	for _, hunk := range hunks {
		for _, line := range hunk {
			switch line.Kind {
			case LineKindAdd, LineKindDel:
				if withLineNumber {
					fmt.Fprintf(w, "%4d:%c%s\n", line.Num, line.Kind, line.Val)
				} else {
					fmt.Fprintf(w, "%c%s\n", line.Kind, line.Val)
				}
			case LineKindLabel:
				w.Write(line.Val)
				w.Write([]byte("\n"))
			default:
				if withLineNumber {
					fmt.Fprintf(w, "%4d: %s\n", line.Num, line.Val)
				} else {
					fmt.Fprintf(w, " %s\n", line.Val)
				}
			}
		}
	}
	return nil
}
