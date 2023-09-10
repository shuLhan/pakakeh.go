// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"fmt"
	"testing"
)

// Test writing to mlog after closed.
func TestMultiLogger_Close(_ *testing.T) {
	var (
		bufOut = bytes.Buffer{}
		bufErr = bytes.Buffer{}
		outs   = []NamedWriter{
			NewNamedWriter("bufOut", &bufOut),
		}
		errs = []NamedWriter{
			NewNamedWriter("bufErr", &bufErr),
		}
		mlog = NewMultiLogger("", "test", outs, errs)

		outq = make(chan struct{})
		errq = make(chan struct{})
	)

	go func() {
		var x int
		for x = 0; x < 10; x++ {
			mlog.Outf("out: %d", x)
			if x == 2 {
				outq <- struct{}{}
				<-outq
			}
		}
	}()
	go func() {
		var x int
		for x = 0; x < 10; x++ {
			mlog.Errf("err: %d", x)
			if x == 2 {
				errq <- struct{}{}
				<-errq
			}
		}
	}()

	<-outq
	<-errq
	mlog.Close()
	outq <- struct{}{}
	errq <- struct{}{}
	mlog.Flush()

	fmt.Println(bufOut.String())
	fmt.Println(bufErr.String())
}
