// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

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
		}
	})
}
