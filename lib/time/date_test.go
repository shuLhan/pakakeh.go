// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package time

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"kilabit.info/pakakeh.go/lib/test"
)

func TestDate_Scan(t *testing.T) {
	got := CreateDate(2026, 1, 1, time.UTC)
	listCase := []struct {
		input    any
		exp      Date
		desc     string
		expError string
	}{{
		desc:     `asString`,
		input:    `2026-04-26`, // Should be []byte not string.
		expError: `time: Date.Scan: invalid type string`,
	}, {
		desc:  `asNull`,
		input: nil,
	}, {
		desc:     `invalidValue`,
		input:    []byte(`2026-1-2`),
		expError: `parsing time "2026-1-2" as "2006-01-02": cannot parse "1-2" as "01"`,
	}, {
		desc:  `OK`,
		input: []byte(`2026-04-26`),
		exp:   CreateDate(2026, 4, 26, time.UTC),
	}}
	for _, tc := range listCase {
		err := got.Scan(tc.input)
		if err != nil {
			test.Assert(t, tc.desc, tc.expError, err.Error())
			continue
		}
		test.Assert(t, tc.desc, tc.exp, got)
	}
}

func TestDate_UnmarshalText(t *testing.T) {
	type sample struct {
		Date Date `json:"date"`
	}
	listCase := []struct {
		exp      sample
		input    string
		expError string
	}{{
		input: `{"date":""}`,
	}, {
		input: `{"date":null}`,
	}, {
		input:    `{"date":"2026-1-2"}`,
		expError: `parsing time "2026-1-2" as "2006-01-02": cannot parse "1-2" as "01"`,
	}, {
		input: `{"date":"2026-04-26"}`,
		exp: sample{
			Date: CreateDate(2026, 4, 26, time.UTC),
		},
	}}
	for _, tc := range listCase {
		s := sample{}
		err := json.Unmarshal([]byte(tc.input), &s)
		if err != nil {
			test.Assert(t, tc.input, tc.expError, err.Error())
			continue
		}
		test.Assert(t, tc.input, tc.exp, s)
	}
}

func TestDate_Value(t *testing.T) {
	listCase := []struct {
		exp  any
		date Date
		desc string
	}{{
		desc: `zero`,
	}, {
		desc: `OK`,
		date: CreateDate(2026, 4, 26, time.UTC),
		exp:  `2026-04-26`,
	}}
	for _, tc := range listCase {
		got, err := tc.date.Value()
		if err != nil {
			log.Fatal(err)
		}
		test.Assert(t, tc.desc, tc.exp, got)
	}
}
