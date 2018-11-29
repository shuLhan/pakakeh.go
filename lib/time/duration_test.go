// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in  string
		exp time.Duration
	}{{
		in:  "1w",
		exp: time.Duration(1) * Week,
	}, {
		in:  "1w1w",
		exp: time.Duration(2) * Week,
	}, {
		in:  "1w0.5w",
		exp: time.Duration(1.5 * float64(Week)),
	}, {
		in:  "1w1d",
		exp: 1*Week + 1*Day,
	}, {
		in:  "0.5d",
		exp: time.Duration(12) * time.Hour,
	}}

	for _, c := range cases {
		t.Log("ParseDuration:", c.in)

		got, err := ParseDuration(c.in)
		if err != nil {
			t.Log(err)
		}

		test.Assert(t, "duration", c.exp, got, true)
	}
}
