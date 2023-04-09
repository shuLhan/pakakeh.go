// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"errors"
	"strconv"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

const (
	// Day the duration for a day.
	Day = 24 * time.Hour
	// Week the duration for a week.
	Week = 7 * Day
)

var (
	// ErrDurationMissingValue an error when value is missing when parsing
	// duration.
	ErrDurationMissingValue = errors.New("missing value in duration")
)

// ParseDuration extend the capability of standard time.Duration with
// additional unit suffix: day and week.
// Day unit end with "d" and week unit end with "w".
// A day is equal with "24h", an a week is equal to "7d".
// Unlike standard time.Duration the week or day units must be before hours.
func ParseDuration(s string) (time.Duration, error) {
	var (
		delims = []byte{' ', '\t', 'w', 'd', 'h', 'm', 's', 'u', 'n'}
		parser = libbytes.NewParser([]byte(s), delims)

		stok string
		tok  []byte
		dur  float64
		v    float64
		c    byte
		err  error
	)

	for {
		tok, c = parser.Read()
		if len(tok) == 0 {
			break
		}

		stok = string(tok)

		v, err = strconv.ParseFloat(stok, 64)
		if err != nil {
			return 0, err
		}

		switch c {
		case 'w':
			dur += v * float64(Week)
		case 'd':
			dur += v * float64(Day)
		default:
			if c != 0 {
				stok += string(c)
			}
			rest, err := time.ParseDuration(stok)
			if err != nil {
				return 0, err
			}
			dur += float64(rest)
		}
	}

	return time.Duration(dur), nil
}
