// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package test provide library for helping with testing.
//
package test

import (
	"runtime"
	"testing"

	"github.com/shuLhan/share/lib/reflect"
)

func printStackTrace(t testing.TB, trace []byte) {
	var (
		lines = 0
		start = 0
		end   = 0
	)
	for x, b := range trace {
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

	t.Log("\n!!! ERR " + string(trace[start:end]))
}

//
// Assert will compare two interfaces: `exp` and `got` for equality.
// If both are not equal, the test will throw panic parameter describe the
// position (type and value) where both are not matched.
//
// If `exp` implement the extended `reflect.Equaler`, then it will use the
// method `IsEqual()` with `got` as parameter.
//
// If debug parameter is true it will print the stack trace of testing.T
// instance.
//
// WARNING: this method does not support recursive pointer, for example node
// that point to parent and parent that point back to node.
//
func Assert(t *testing.T, name string, exp, got interface{}, debug bool) {
	err := reflect.DoEqual(exp, got)
	if err == nil {
		return
	}

	if debug {
		trace := make([]byte, 1024)
		runtime.Stack(trace, false)
		printStackTrace(t, trace)
	}

	t.Fatalf("!!! %s: %s", name, err)
}

//
// AssertBench will compare two interfaces: `exp` and `got` whether its same
// with `equal` value.
//
// If comparison result is not same with `equal`, it will print the result and
// expectation and then terminate the test routine.
//
func AssertBench(b *testing.B, name string, exp, got interface{}, equal bool) {
	if reflect.IsEqual(exp, got) == equal {
		return
	}

	trace := make([]byte, 1024)
	runtime.Stack(trace, false)

	printStackTrace(b, trace)

	b.Fatalf(">>> Got %s:\n\t'%+v';\n"+
		"     want:\n\t'%+v'\n", name, got, exp)
}

func AssertBench2(b *testing.B, name string, exp, got interface{}) {
	err := reflect.DoEqual(exp, got)
	if err == nil {
		return
	}

	trace := make([]byte, 1024)
	runtime.Stack(trace, false)

	printStackTrace(b, trace)

	b.Fatalf("!!! %s: %s", name, err)
}
