// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package test provide library for helping with testing.
package test

import (
	"runtime"

	"github.com/shuLhan/share/lib/reflect"
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

	_, ok = w.(*testWriter)
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
// will use the text/diff.Text to show the difference between them.
// The diff output is as follow,
//
//	!!! "string not matched" / <desc>:
//	----
//	<LINE_NUM> - "<STRING>"
//	...
//	++++
//	<LINE_NUM> + "<STRING>"
//	...
//	--++
//	<LINE_NUM> - "<LINE_EXP>"
//	<LINE_NUM> + "<LINE_GOT>"
//	^<COL_NUM> - "<DELETED_STRING>"
//	^<COL_NUM> + "<INSERTED_STRING>"
//
// Any lines after "----" indicate the lines that deleted in got (exist in exp
// but not in got).
//
// Any lines after "++++" indicate the lines that inserted in got (does not
// exist in exp but exist in got).
//
// Any lines after "--++" indicate that the line between exp and got has words
// changes in it.
//
//   - The "<LINE_NUM> - " print the line in exp.
//   - The "<LINE_NUM> + " print the line in got.
//   - The "^<COL_NUM> - " print the position and the string deleted in exp
//     (or string that not exist in got).
//   - The "^<COL_NUM> + " print the position and the string inserted in got
//     (or string that not exist in exp).
//
// WARNING: this method does not support recursive pointer, for example a node
// that point to parent and parent that point back to node again.
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

	if len(name) == 0 {
		w.Fatal("!!! strings not matched:\n", diffData.String())
	} else {
		w.Fatalf("!!! %s:\n%s", name, diffData.String())
	}

	return true
}
