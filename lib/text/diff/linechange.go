// Copyright 2018 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

	Old text.Line
	New text.Line
}

// NewLineChange create a pointer to new LineChange object.
func NewLineChange(old, new text.Line) *LineChange {
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

	fmt.Fprintf(&sb, "%d - %q\n", change.Old.N, change.Old.V)
	fmt.Fprintf(&sb, "%d + %q\n", change.New.N, change.New.V)

	for _, chunk = range change.Dels {
		fmt.Fprintf(&sb, "^%d - %q\n", chunk.StartAt, chunk.V)
	}
	for _, chunk = range change.Adds {
		fmt.Fprintf(&sb, "^%d + %q\n", chunk.StartAt, chunk.V)
	}
	return sb.String()
}
