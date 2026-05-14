// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

// Package test provide library for helping with testing.
package test

import (
	"bytes"

	"kilabit.info/pakakeh.go/lib/reflect"
	"kilabit.info/pakakeh.go/lib/text/diff"
)

// Assert compares two interfaces values, exp and got for equality.
// If both parameters are not equal, the function will try to describe the
// type and value for both of them and print it with [testing.T.Fatalf].
//
// If exp implement the extended [reflect.Equaler], then it will use the
// method [Equal] and pass got as parameter.
//
// If exp and got is a struct, it will print the first non-matching field in
// the following format,
//
//	!!! Assert: <name>: T.<Field>: expecting <type>(<value>), got <type>(<value>)
//
// If both exp and got types are string and its longer than 50 chars, it
// will use the [diff.Unified] to show the difference between them.
//
//	!!! Assert: <name>:
//	<UnifiedDiff>
//
// LIMITATION: this method does not support recursive pointer, for example a
// node that point to parent and parent that point back to node again.
func Assert(w Writer, name string, exp, got any) {
	w.Helper()

	err := reflect.DoEqual(exp, got)
	if err == nil {
		return
	}

	logp := `Assert`
	var gotStr string
	var diffd *diff.Data
	var bb bytes.Buffer

	expStr, ok := exp.(string)
	if !ok {
		goto out
	}
	gotStr, ok = got.(string)
	if !ok {
		goto out
	}
	if max(len(expStr), len(gotStr)) < 50 {
		goto out
	}

	diffd = diff.Unified([]byte(expStr), []byte(gotStr))
	if diffd.IsMatched {
		return
	}
	diffd.OldName = `want`
	diffd.NewName = `got`
	diffd.WriteUnified(&bb, 3)
	w.Fatalf("!!! %s: %s: not matched, see the diff below\n%s", logp, name, bb.String())
	return

out:
	if len(name) == 0 {
		w.Fatalf(`!!! %s: %s`, logp, err)
	} else {
		w.Fatalf(`!!! %s: %s: %s`, logp, name, err)
	}
}
