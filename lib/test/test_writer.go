// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bytes"
	"fmt"
)

// testWriter implement some of testing.TB.
type testWriter struct {
	bytes.Buffer
}

func (tw *testWriter) Error(args ...any)                 {}
func (tw *testWriter) Errorf(format string, args ...any) {}
func (tw *testWriter) Fatal(args ...any)                 { fmt.Fprint(tw, args...) }
func (tw *testWriter) Fatalf(format string, args ...any) { fmt.Fprintf(tw, format, args...) }
func (tw *testWriter) Log(args ...any)                   { fmt.Fprint(tw, args...) }
func (tw *testWriter) Logf(format string, args ...any)   { fmt.Fprintf(tw, format, args...) }
