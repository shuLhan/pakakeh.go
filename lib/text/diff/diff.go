// Copyright 2018 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package diff implement text comparison.
//
package diff

import (
	"fmt"

	"github.com/shuLhan/share/lib/text"
)

var (
	// DefDelimiter define default delimiter for new line.
	DefDelimiter = byte('\n') //nolint: gochecknoglobals
)

//
// Data represent additions, deletions, and changes between two text.
//
type Data struct {
	Adds    text.Lines
	Dels    text.Lines
	Changes LineChanges
}

//
// PushAdd will add new line to diff set.
//
func (diffs *Data) PushAdd(new text.Line) {
	diffs.Adds = append(diffs.Adds, new)
}

//
// PushDel will add deletion line to diff set.
//
func (diffs *Data) PushDel(old text.Line) {
	diffs.Dels = append(diffs.Dels, old)
}

//
// PushChange set to diff data.
//
func (diffs *Data) PushChange(old, new text.Line) {
	change := NewLineChange(old, new)

	diffs.Changes = append(diffs.Changes, *change)
}

//
// GetAllAdds return chunks of additions including in line changes.
//
func (diffs *Data) GetAllAdds() (chunks text.Chunks) {
	for _, add := range diffs.Adds {
		chunks = append(chunks, text.Chunk{StartAt: 0, V: add.V})
	}
	chunks = append(chunks, diffs.Changes.GetAllAdds()...)
	return
}

//
// GetAllDels return chunks of deletions including in line changes.
//
func (diffs *Data) GetAllDels() (chunks text.Chunks) {
	for _, del := range diffs.Dels {
		chunks = append(chunks, text.Chunk{StartAt: 0, V: del.V})
	}
	chunks = append(chunks, diffs.Changes.GetAllDels()...)
	return
}

//
// String return formatted data.
//
func (diffs Data) String() (s string) {
	s += "Diffs:\n"

	if len(diffs.Adds) > 0 {
		s += ">>> Adds:\n"
		for _, add := range diffs.Adds {
			s += fmt.Sprintf("  + %d : %s", add.N, string(add.V))
		}
	}

	if len(diffs.Dels) > 0 {
		s += ">>> Dels:\n"
		for _, del := range diffs.Dels {
			s += fmt.Sprintf("  - %d : %s", del.N, string(del.V))
		}
	}

	if len(diffs.Changes) > 0 {
		s += ">>> Changes:\n" + fmt.Sprint(diffs.Changes)
	}

	return
}
