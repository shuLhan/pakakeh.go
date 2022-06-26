// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"testing"
)

func BenchmarkMultiLogger(b *testing.B) {
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
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mlog.Errf("err")
			mlog.Outf("out")
			//mlog.Flush()
		}
	})
}
