// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package diff

import (
	"bytes"
	"fmt"
	"os"
)

// Line represents single line from file or stream of bytes.
type Line struct {
	Val []byte // The line value.
	Num int    // The line number, started from 1.
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
		lines = append(lines, Line{Num: x + 1, Val: line})
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
