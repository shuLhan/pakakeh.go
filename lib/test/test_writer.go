// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bytes"
	"fmt"
)

// TestWriter write Errorx, Fatalx, and Logx to bytes.Buffer.
type TestWriter struct {
	bytes.Buffer
}

func (tw *TestWriter) Error(args ...any)                 {}
func (tw *TestWriter) Errorf(format string, args ...any) {}
func (tw *TestWriter) Fatal(args ...any)                 { fmt.Fprint(tw, args...) }
func (tw *TestWriter) Fatalf(format string, args ...any) { fmt.Fprintf(tw, format, args...) }
func (tw *TestWriter) Log(args ...any)                   { fmt.Fprint(tw, args...) }
func (tw *TestWriter) Logf(format string, args ...any)   { fmt.Fprintf(tw, format, args...) }
