// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package time

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Clock represent 24 hours time with hour, minute, and second.
// An hour value is from 0 to 23, a minute value is from 0 to 59, and a second
// value is from 0 to 59.
type Clock struct {
	hour int
	min  int
	sec  int
}

// CreateClock create instance of Clock.
// Any value that is not valid will be set to 0.
func CreateClock(hour, min, sec int) Clock {
	if hour < 0 || hour > 23 {
		hour = 0
	}
	if min < 0 || min > 59 {
		min = 0
	}
	if sec < 0 || sec > 59 {
		sec = 0
	}
	return Clock{
		hour: hour,
		min:  min,
		sec:  sec,
	}
}

// ParseClock parse the clock from string with format `HH:MM:SS`.
// The MM and SS are optionals.
// Any value that is not valid will be set to 0.
func ParseClock(v string) (c Clock) {
	var (
		vals = strings.Split(v, `:`)

		err error
	)

	c.hour, err = strconv.Atoi(vals[0])
	if err == nil {
		if c.hour < 0 || c.hour > 23 {
			c.hour = 0
		}
	}

	if len(vals) >= 2 {
		c.min, err = strconv.Atoi(vals[1])
		if err == nil {
			if c.min < 0 || c.min > 59 {
				c.min = 0
			}
		}
	}
	if len(vals) >= 3 {
		c.sec, err = strconv.Atoi(vals[2])
		if err == nil {
			if c.sec < 0 || c.sec > 59 {
				c.sec = 0
			}
		}
	}
	return c
}

// SortClock sort the clock.
func SortClock(list []Clock) {
	sort.SliceStable(list, func(x, y int) bool {
		return list[x].Before(list[y])
	})
}

// After return true if the Clock instant c is after d.
func (c Clock) After(d Clock) bool {
	if c.hour > d.hour {
		return true
	}
	if c.hour < d.hour {
		return false
	}
	// hour==hour
	if c.min > d.min {
		return true
	}
	if c.min < d.min {
		return false
	}
	// minute==minute
	if c.sec > d.sec {
		return true
	}
	// An equal second is not an After.
	return false
}

// Before return true if the Clock instant c is before d.
func (c Clock) Before(d Clock) bool {
	if c.hour > d.hour {
		return false
	}
	if c.hour < d.hour {
		return true
	}
	// hour==hour
	if c.min > d.min {
		return false
	}
	if c.min < d.min {
		return true
	}
	// minute==minute
	if c.sec >= d.sec {
		return false
	}
	return true
}

// Equal return true if the Clock instance c is equal with d.
func (c Clock) Equal(d Clock) bool {
	if c.hour != d.hour {
		return false
	}
	if c.min != d.min {
		return false
	}
	return c.sec == d.sec
}

// Hour return the Clock hour value.
func (c Clock) Hour() int {
	return c.hour
}

// Minute return the Clock minute value.
func (c Clock) Minute() int {
	return c.min
}

// Second return the Clock second value.
func (c Clock) Second() int {
	return c.sec
}

// String return the Clock value as "HH:MM:SS".
func (c Clock) String() string {
	return fmt.Sprintf(`%02d:%02d:%02d`, c.hour, c.min, c.sec)
}
