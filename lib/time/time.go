// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package time provide a library for working with time.
package time

import (
	"strconv"
	"time"
)

// timeNow returns the current local time.
// This is a variable that can be override to mock the current time during
// testing.
var timeNow = time.Now

// ShortDayNames contains list of day name in English, in shorter.
var ShortDayNames = []string{`Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`, `Sun`}

// ShortMonths provide mapping between text of month, in English,
// short format to their time.Month value
var ShortMonths = map[string]time.Month{
	`Jan`: time.January,
	`Feb`: time.February,
	`Mar`: time.March,
	`Apr`: time.April,
	`May`: time.May,
	`Jun`: time.June,
	`Jul`: time.July,
	`Aug`: time.August,
	`Sep`: time.September,
	`Oct`: time.October,
	`Nov`: time.November,
	`Dec`: time.December,
}

// Microsecond return the microsecond value of time.
// For example, if the unix nano seconds is 1612331218913557000 then the micro
// second value is 913557.
//
// To get the unix microsecond use [time.Time.UnixMicro].
func Microsecond(t *time.Time) int64 {
	seconds := t.Unix() * int64(time.Second)
	return (t.UnixNano() - seconds) / int64(time.Microsecond)
}

// UnixMilliString returns the UnixMilli() as string.
func UnixMilliString(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}
