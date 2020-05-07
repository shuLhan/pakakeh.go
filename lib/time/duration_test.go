// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/strings"
	"github.com/shuLhan/share/lib/test"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in     string
		expErr []string
		exp    time.Duration
	}{{
		in: "w",
		expErr: []string{
			ErrDurationMissingValue.Error(),
		},
	}, {
		in: "1aw",
		expErr: []string{
			`strconv.ParseFloat: parsing "1a": invalid syntax`,
		},
	}, {
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
	}, {
		in:  "1d0.5h",
		exp: time.Duration(24)*time.Hour + (time.Minute * time.Duration(30)),
	}, {
		in: "100  w",
		expErr: []string{
			`time: unknown unit   in duration 100 `,     // go <= 1.14
			`time: unknown unit " " in duration "100 "`, // go > 1.14
		},
	}, {
		in: "100",
	}}

	for _, c := range cases {
		t.Log("ParseDuration:", c.in)

		got, err := ParseDuration(c.in)
		if err != nil {
			if strings.IsContain(c.expErr, err.Error()) {
				continue
			}
			test.Assert(t, "error", c.expErr, err.Error(), true)
		}

		test.Assert(t, "duration", c.exp, got, true)
	}
}
