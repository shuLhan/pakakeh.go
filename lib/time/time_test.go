// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestMicrosecond(t *testing.T) {
	now := time.Now()
	seconds := now.Unix()
	seconds0 := seconds * int64(time.Second)
	nanos := now.UnixNano()
	micros := Microsecond(&now)
	t.Logf("Seconds    : %d\n", seconds)
	t.Logf("Seconds 0  : %d\n", seconds0)
	t.Logf("Nanosecond : %d\n", nanos)
	t.Logf("Microsecond: %d\n", micros)
}

func TestUnixMilliString(t *testing.T) {
	ms := int64(123 * time.Millisecond)
	tt := time.Unix(1000000000, ms)

	exp := "1000000000123"
	got := UnixMilliString(tt)

	test.Assert(t, "UnixMilliString", exp, got, true)
}
