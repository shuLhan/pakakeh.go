// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package time provide a library for working with time.
package time

import (
	"time"
)

//nolint:gochecknoglobals
var (
	ShortDayNames = []string{
		"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun",
	}

	//
	// ShortMonths provide mapping between text of month, in English,
	// short format to their time.Month value
	//
	ShortMonths = map[string]time.Month{
		"Jan": time.January,
		"Feb": time.February,
		"Mar": time.March,
		"Apr": time.April,
		"May": time.May,
		"Jun": time.June,
		"Jul": time.July,
		"Aug": time.August,
		"Sep": time.September,
		"Oct": time.October,
		"Nov": time.November,
		"Dec": time.December,
	}
)

//
// Microsecond return the microsecond value of time.
//
func Microsecond(t *time.Time) int {
	seconds := t.Unix() * int64(time.Second)
	return int(t.UnixNano()-seconds) / int(time.Microsecond)
}
