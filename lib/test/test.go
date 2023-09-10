// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package test provide library for helping with testing.
package test

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/shuLhan/share/lib/reflect"
	"github.com/shuLhan/share/lib/text"
	"github.com/shuLhan/share/lib/text/diff"
)

func printStackTrace(w Writer, trace []byte) {
	var (
		lines int
		start int
		end   int
		x     int
		b     byte
		ok    bool
	)

	for x, b = range trace {
		if b == '\n' {
			lines++
			if lines == 3 {
				start = x + 1
			}
			if lines == 5 {
				end = x + 1
				break
			}
		}
	}

	_, ok = w.(*BufferWriter)
	if !ok {
		w.Log("\n!!! ERR " + string(trace[start:end]))
	}
}

// Assert compare two interfaces: exp and got for equality.
// If both parameters are not equal, the function will call print and try to
// describe the position (type and value) where value are not matched and call
// Fatalf.
//
// If exp implement the extended reflect.Equaler, then it will use the
// method IsEqual with got as parameter.
//
// If exp and got is a struct, it will print the first non-matched field in
// the following format,
//
//	!!! Assert: [<name>: ] T.<Field>: expecting <type>(<value>), got <type>(<value>)
//
// If both exp and got types are string and its longer than 50 chars, it
// will use the [diff.Text] to show the difference between them.
// The diff output is as follow,
//
//	!!! <name>:
//	---- EXPECTED
//	<LINE_NUM> - "<STRING>"
//	...
//	++++ GOT
//	<LINE_NUM> + "<STRING>"
//	...
//	--++
//	<LINE_NUM> - "<LINE_EXP>"
//	<LINE_NUM> + "<LINE_GOT>"
//
// Any lines after "----" indicate the lines that test expected, from `exp`
// parameter.
//
// Any lines after "++++" indicate the lines that test got, from `got`
// parameter.
//
// Any lines after "--++" indicate that the same line between expected and got
// but different content.
//
//   - The "<LINE_NUM> - " print the expected line.
//   - The "<LINE_NUM> + " print the got line.
//
// LIMITATION: this method does not support recursive pointer, for example a
// node that point to parent and parent that point back to node again.
func Assert(w Writer, name string, exp, got interface{}) {
	var (
		logp = `Assert`

		err   error
		trace []byte
	)

	err = reflect.DoEqual(exp, got)
	if err == nil {
		return
	}

	if printStringDiff(w, name, exp, got) {
		return
	}

	trace = make([]byte, 1024)
	runtime.Stack(trace, false)
	printStackTrace(w, trace)

	if len(name) == 0 {
		w.Fatalf(`!!! %s: %s`, logp, err)
	} else {
		w.Fatalf(`!!! %s: %s: %s`, logp, name, err)
	}
}

func printStringDiff(w Writer, name string, exp, got interface{}) bool {
	var (
		diffData diff.Data
		expStr   string
		gotStr   string
		ok       bool
	)

	expStr, ok = exp.(string)
	if !ok {
		return false
	}

	gotStr, ok = got.(string)
	if !ok {
		return false
	}

	if len(expStr) < 50 {
		return false
	}

	diffData = diff.Text([]byte(expStr), []byte(gotStr), diff.LevelWords)
	if diffData.IsMatched {
		return true
	}

	var (
		bb   bytes.Buffer
		line text.Line
	)

	fmt.Fprintf(&bb, "!!! %s:\n", name)

	if len(diffData.Dels) > 0 {
		bb.WriteString("---- EXPECTED\n")
		for _, line = range diffData.Dels {
			fmt.Fprintf(&bb, "%d - %s\n", line.N, line.V)
		}
	}

	if len(diffData.Adds) > 0 {
		bb.WriteString("++++ GOT\n")
		for _, line = range diffData.Adds {
			fmt.Fprintf(&bb, "%d + %s\n", line.N, line.V)
		}
	}

	if len(diffData.Changes) > 0 {
		bb.WriteString("--++\n")

		var change diff.LineChange
		for _, change = range diffData.Changes {
			fmt.Fprintf(&bb, "%d - %s\n", change.Old.N, change.Old.V)
			fmt.Fprintf(&bb, "%d + %s\n", change.New.N, change.New.V)
		}
	}

	w.Fatal(bb.String())

	return true
}
