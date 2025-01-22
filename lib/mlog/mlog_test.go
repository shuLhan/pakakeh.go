// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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
		for x = range 10 {
			mlog.Outf("out: %d", x)
			if x == 2 {
				outq <- struct{}{}
				<-outq
			}
		}
	}()
	go func() {
		var x int
		for x = range 10 {
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
