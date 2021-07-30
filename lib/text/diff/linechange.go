// Copyright 2018 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"fmt"

	"github.com/shuLhan/share/lib/text"
)

//
// LineChange represent one change in text.
//
type LineChange struct {
	Old  text.Line
	New  text.Line
	Adds text.Chunks
	Dels text.Chunks
}

//
// NewLineChange create a pointer to new LineChange object.
//
func NewLineChange(old, new text.Line) *LineChange {
	return &LineChange{old, new, text.Chunks{}, text.Chunks{}}
}

//
// String return formatted content of LineChange.
//
func (change LineChange) String() string {
	return fmt.Sprintf("LineChange: {\n"+
		" Old  : %s\n"+
		" New  : %s\n"+
		" Adds : %s\n"+
		" Dels : %s\n"+
		"}\n", change.Old, change.New, change.Adds, change.Dels)
}
