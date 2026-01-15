// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package text

// Lines represent array of line.
type Lines []Line

// ParseLines convert raw bytes into Lines.
func ParseLines(raw []byte) (lines Lines) {
	var (
		start   int
		lineNum int
	)
	for x, c := range raw {
		if c == '\n' {
			line := Line{
				N: lineNum,
				V: raw[start:x],
			}
			lines = append(lines, line)
			start = x + 1
			lineNum++
		}
	}
	if start < len(raw) {
		line := Line{
			N: lineNum,
			V: raw[start:],
		}
		lines = append(lines, line)
	}
	return lines
}
