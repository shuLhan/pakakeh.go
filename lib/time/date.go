// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package time

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// A Date represents the calendar date.
// This type can be used to represent PostgreSQL date.
// It provides encodes and decodes for [encoding/json] and [database/sql] with
// fixed format [time.DateOnly] or "2006-01-02".
type Date struct {
	v time.Time
}

// CreateDate returns the calendar date corresponding to the year, month, day,
// and location.
func CreateDate(year int, month time.Month, day int, loc *time.Location) Date {
	d := Date{
		v: time.Date(year, month, day, 0, 0, 0, 0, loc),
	}
	return d
}

// IsZero returns true if date is not set.
func (date Date) IsZero() bool {
	return date.v.IsZero()
}

// MarshalText encodes the date as string.
// This method support marshaling into JSON with [encoding/json] package.
func (date Date) MarshalText() ([]byte, error) {
	if date.v.IsZero() {
		return nil, nil
	}
	return []byte(date.String()), nil
}

// Scan scans the value src into Date.
// This method implements the [database/sql.Scanner] interface.
func (date *Date) Scan(src any) (err error) {
	switch v := src.(type) {
	case time.Time:
		date.v = v
	case []byte:
		err = date.UnmarshalText(v)
		if err != nil {
			return err
		}
	case nil:
		date.v = time.Time{}
		return nil
	default:
		return fmt.Errorf(`time: Date.Scan: invalid type %T`, v)
	}
	return nil
}

// String returns the date as string.
func (date *Date) String() string {
	return date.v.Format(time.DateOnly)
}

// UnmarshalText decodes the stream of bytes into date.
// This method support unmarshaling from JSON string.
func (date *Date) UnmarshalText(v []byte) (err error) {
	if len(v) == 0 {
		date.v = time.Time{}
		return nil
	}
	date.v, err = time.Parse(time.DateOnly, string(v))
	if err != nil {
		return err
	}
	return nil
}

// Value returns the value of date as database value.
// This method implements the [driver.Valuer] interface.
// It will return nil if the date value is zero.
func (date Date) Value() (v driver.Value, err error) {
	if date.v.IsZero() {
		return nil, nil
	}
	v = date.String()
	return v, nil
}
