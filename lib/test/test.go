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
//	!!! string not matched:
//	--++
//	<LINE_NUM> - "<LINE_EXP>"
//	<LINE_NUM> + "<LINE_GOT>"
//	^<COL_NUM> - "<DELETED_STRING>"
//	^<COL_NUM> + "<INSERTED_STRING>"
//
// The "<LINE_NUM> - " print the line number in exp followed by line itself.
// The "<LINE_NUM> + " print the line number in got followed by line itself.
// The "^<COL_NUM> - " show the character number in exp line followed by
// deleted string (or string that not exist in got).
// The "^<COL_NUM> + " show the character number in got line followed by
// inserted string (or string that not exist in exp).
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

	trace = make([]byte, 1024)
	runtime.Stack(trace, false)
	printStackTrace(w, trace)

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

	if len(name) == 0 {
		w.Log("!!! strings not matched:\n", diffData.String())
	} else {
		w.Logf("!!! %s:\n%s", name, diffData.String())
	}

	return true
}
