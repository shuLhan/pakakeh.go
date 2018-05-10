// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package test provide library for help with testing.
//
package test

import (
	"os"
	"reflect"
	"runtime"
	"testing"
)

var (
	trace = make([]byte, 1024)
)

func printStackTrace(t testing.TB) {
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
// If comparison result is not same with `equal`, it will terminate the test
// program.
//
func Assert(t *testing.T, name string, exp, got interface{}, equal bool) {
	if reflect.DeepEqual(exp, got) != equal {
		runtime.Stack(trace, false)

		printStackTrace(t)

		t.Fatalf(">>> Expecting %s,\n"+
			"'%+v'\n"+
			"     got,\n"+
			"'%+v'\n", name, exp, got)
		os.Exit(1)
	}
}

//
// AssertBench will compare two interfaces: `exp` and `got` whether its same
// with `equal` value.
//
// If comparison result is not same with `equal`, it will terminate the test
// program.
//
func AssertBench(b *testing.B, name string, exp, got interface{}, equal bool) {
	if reflect.DeepEqual(exp, got) != equal {
		runtime.Stack(trace, false)

		printStackTrace(b)

		b.Fatalf("\n"+
			">>> Expecting %s '%+v'\n"+
			"    got '%+v'\n", name, exp, got)
		os.Exit(1)
	}
}
