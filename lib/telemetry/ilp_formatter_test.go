// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestIlpFormatter_Format(t *testing.T) {
	var md = NewMetadata()
	md.Set(`host`, `localhost`)
	md.Set(`version`, `0.1.0`)

	var ilp = NewIlpFormatter(`myapp`)

	test.Assert(t, `Name`, `ilp`, ilp.Name())

	var (
		m   Metric
		got []byte
		exp string
	)

	got = ilp.Format(m, md)
	test.Assert(t, `Format: empty`, exp, string(got))

	m = Metric{
		Timestamp: 1000,
		Name:      `go_gc_total`,
		Value:     0.004,
	}
	got = ilp.Format(m, md)

	exp = `myapp,host=localhost,version=0.1.0 go_gc_total=0.004000 1000`

	test.Assert(t, `Format`, string(exp), string(got))
}

func TestIlpFormatter_BulkFormat(t *testing.T) {
	var md = NewMetadata()
	md.Set(`host`, `localhost`)
	md.Set(`version`, `0.1.0`)

	var ilp = NewIlpFormatter(`myapp`)

	test.Assert(t, `Name`, `ilp`, ilp.Name())

	var (
		list = []Metric{}
		got  = ilp.BulkFormat(list, md)

		exp string
	)

	test.Assert(t, `BulkFormat: empty`, exp, string(got))

	list = append(list, Metric{
		Timestamp: 1000,
		Name:      `go_gc_total`,
		Value:     0.004,
	})
	list = append(list, Metric{
		Timestamp: 1000,
		Name:      `go_gc_pause_seconds`,
		Value:     0.00001,
	})

	got = ilp.BulkFormat(list, md)

	exp = "myapp,host=localhost,version=0.1.0 go_gc_total=0.004000,go_gc_pause_seconds=0.000010 1000\n"

	test.Assert(t, `BulkFormat`, string(exp), string(got))
}
