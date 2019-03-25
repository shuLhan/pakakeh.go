// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package test provide library for helping with testing.
//
package test

import (
	"reflect"
	"runtime"
	"testing"
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
// Assert will compare two interfaces: `exp` and `got` whether its same with
// `equal` value.
//
// If comparison result is not same with `equal`, it will print the result and
// expectation and then terminate the test routine.
//
func Assert(t *testing.T, name string, exp, got interface{}, equal bool) {
	if reflect.DeepEqual(exp, got) == equal {
		return
	}

	trace := make([]byte, 1024)
	runtime.Stack(trace, false)

	printStackTrace(t, trace)

	t.Fatalf(">>> Got %s:\n\t'%+v';\n"+
		"     want:\n\t'%+v'\n", name, got, exp)
}

//
// AssertBench will compare two interfaces: `exp` and `got` whether its same
// with `equal` value.
//
// If comparison result is not same with `equal`, it will print the result and
// expectation and then terminate the test routine.
//
func AssertBench(b *testing.B, name string, exp, got interface{}, equal bool) {
	if reflect.DeepEqual(exp, got) == equal {
		return
	}

	trace := make([]byte, 1024)
	runtime.Stack(trace, false)

	printStackTrace(b, trace)

	b.Fatalf(">>> Got %s:\n\t'%+v';\n"+
		"     want:\n\t'%+v'\n", name, got, exp)
}
