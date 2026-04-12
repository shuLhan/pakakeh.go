// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package diff

import (
	"fmt"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
)

// LineChunk represents a changes inside a line.
// Each chunk in Adds and Dels contains the position and value of string that
// is added in New or deleted in Old line.
type LineChunk struct {
	Adds text.Chunks
	Dels text.Chunks

	Old Line
	New Line
}

// String returns formatted content of LineChunk.
// The format is as follow,
//
//	<OldLine>
//	<NewLine>
//	<StartAt>: <Kind><QuotedVal>
//
// The OldLine and NewLine print the line being compared, following the format
// from [Line.String].
// The StartAt print the beginning position where characters differences found
// between old and new line.
// The Kind print '-' for deletion or '+' for deletion.
// The QuoteVal print the string difference.
func (lchunk LineChunk) String() string {
	var sb strings.Builder

	sb.WriteString(lchunk.Old.String())
	sb.WriteByte('\n')
	sb.WriteString(lchunk.New.String())
	sb.WriteByte('\n')

	for _, chunk := range lchunk.Dels {
		fmt.Fprintf(&sb, "^%-4d: -%q\n", chunk.StartAt, chunk.V)
	}
	for _, chunk := range lchunk.Adds {
		fmt.Fprintf(&sb, "^%-4d: +%q\n", chunk.StartAt, chunk.V)
	}
	return sb.String()
}
