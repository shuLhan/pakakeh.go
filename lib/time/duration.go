// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"errors"
	"strconv"
	"time"

	libio "github.com/shuLhan/share/lib/io"
	libtext "github.com/shuLhan/share/lib/text"
)

const (
	Day  = 24 * time.Hour
	Week = 7 * Day
)

var (
	ErrDurationMissingValue = errors.New("Missing value in duration")
)

//
// ParseDuration extend the capability of standard time.Duration with
// additional unit suffix: day and week.
// Day unit end with "d" and week unit end with "w".
// A day is equal with "24h", an a week is equal to "7d".
// Unlike standard time.Duration the week or day units must be before hours.
//
func ParseDuration(s string) (time.Duration, error) {
	var (
		dur, v float64
		err    error
	)

	seps := []byte{'w', 'd', 'h', 'm', 's', 'u', 'n'}

	reader := &libio.Reader{}
	reader.Init(s)

	c := reader.SkipSpace()
	if !libtext.IsDigit(c) {
		return 0, ErrDurationMissingValue
	}

	for {
		tok, isTerm, c := reader.ReadUntil(seps, libtext.ASCIISpaces)
		if len(tok) == 0 {
			break
		}

		stok := string(tok)

		switch c {
		case 'w':
			v, err = strconv.ParseFloat(stok, 64)
			if err != nil {
				return 0, err
			}
			dur += v * float64(Week)
		case 'd':
			v, err = strconv.ParseFloat(stok, 64)
			if err != nil {
				return 0, err
			}
			dur += v * float64(Day)
		default:
			s := stok + reader.String()
			rest, err := time.ParseDuration(s)
			if err != nil {
				return 0, nil
			}
			dur += float64(rest)
			break
		}
		if isTerm {
			break
		}
	}

	return time.Duration(dur), nil
}
