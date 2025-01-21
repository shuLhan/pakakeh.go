// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package test provide library for helping with testing.
package test

import (
	"bytes"
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
	"git.sr.ht/~shulhan/pakakeh.go/lib/text/diff"
)

// Assert compare two interfaces: exp and got for equality.
// If both parameters are not equal, the function will call print and try to
// describe the position (type and value) where value are not matched and call
// Fatalf.
//
// If exp implement the extended [reflect.Equaler], then it will use the
// method [Equal] with got as parameter.
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
	w.Helper()

	var logp = `Assert`
	var err error

	err = reflect.DoEqual(exp, got)
	if err == nil {
		return
	}

	if printStringDiff(w, name, exp, got) {
		return
	}

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
