// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package diff

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// List of kind of line in [Data].
const (
	LineKindAdd   = '+'
	LineKindDel   = '-'
	LineKindLabel = '@'
	LineKindNone  = 0
)

// Line represents single line from file or stream of bytes.
type Line struct {
	Val      []byte // The line value.
	Kind     int    // The kind of changes.
	Num      int    // The line number, started from 1.
	NumOther int    // The line number of other file.
}

// ParseLines returns lines from reading raw bytes.
// The line number returned in lines started from 1.
func ParseLines(raw []byte) (lines []Line) {
	rawlines := bytes.Split(raw, []byte("\n"))
	lenraw := len(rawlines)
	if len(rawlines[lenraw-1]) == 0 {
		rawlines = rawlines[:lenraw-1]
	}
	lines = make([]Line, 0, len(rawlines))
	for x, line := range rawlines {
		l := Line{Num: x + 1, NumOther: x + 1, Val: line}
		lines = append(lines, l)
	}
	return lines
}

// ReadLines returns lines in the file.
// The line number returned in lines started from 1.
func ReadLines(file string) (lines []Line, err error) {
	logp := `ReadLines`
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	lines = ParseLines(raw)
	return lines, nil
}

// WriteLines writes each line in lines into w using the [Line.String] return
// value.
func WriteLines(w io.Writer, lines []Line) {
	for _, line := range lines {
		fmt.Fprintln(w, line.String())
	}
}

// Strings returns the formatted line data.
// The format is as follow,
//
//	<NumOther>/<Num>: <Kind><QuotedVal>
//
// The NumOther is line number from other text being compared.
// The Num is line number.
// The Kind is '-' for deletion or '+' for addition.
// The QuotedVal is double quoted line text, to help human compare and see the
// changes for non-printable characters.
func (line Line) String() string {
	return fmt.Sprintf("%4d/%4d: %c%q", line.NumOther, line.Num, line.Kind, line.Val)
}
