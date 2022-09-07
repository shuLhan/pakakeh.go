// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package test provide library for helping with testing.
package test

import (
	"runtime"

	"github.com/shuLhan/share/lib/reflect"
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

// Assert compare two interfaces: `exp` and `got` for equality.
// If both parameters are not equal, the function will call Fatalf that
// describe the position (type and value) where value are not matched.
//
// If `exp` implement the extended `reflect.Equaler`, then it will use the
// method `IsEqual()` with `got` as parameter.
//
// WARNING: this method does not support recursive pointer, for example a node
// that point to parent and parent that point back to node again.
func Assert(w Writer, name string, exp, got interface{}) {
	var (
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

	if len(name) == 0 {
		w.Fatalf(`!!! %s`, err)
	} else {
		w.Fatalf(`!!! %s: %s`, name, err)
	}
}
