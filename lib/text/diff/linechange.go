// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package diff

import (
	"fmt"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
)

// LineChange represent one change in text.
type LineChange struct {
	Adds text.Chunks
	Dels text.Chunks

	Old Line
	New Line
}

// NewLineChange create a pointer to new LineChange object.
func NewLineChange(old, new Line) *LineChange {
	return &LineChange{
		Adds: text.Chunks{},
		Dels: text.Chunks{},
		Old:  old,
		New:  new,
	}
}

// String return formatted content of LineChange.
func (change LineChange) String() string {
	var (
		sb    strings.Builder
		chunk text.Chunk
	)

	fmt.Fprintf(&sb, "%d - %q\n", change.Old.Num, change.Old.Val)
	fmt.Fprintf(&sb, "%d + %q\n", change.New.Num, change.New.Val)

	for _, chunk = range change.Dels {
		fmt.Fprintf(&sb, "^%d - %q\n", chunk.StartAt, chunk.V)
	}
	for _, chunk = range change.Adds {
		fmt.Fprintf(&sb, "^%d + %q\n", chunk.StartAt, chunk.V)
	}
	return sb.String()
}
